package main

import (
	"encoding/json"
	"fmt"
	"github.com/gorilla/mux"
	"github.com/nu7hatch/gouuid"
	"html/template"
	"io"
	"net/http"
	"os"
	"strings"
)

var (
	maxBytes = int64(1024 * 5000)
)

type Service struct {
	r *mux.Router
	t TreasureHunts
}

func (s *Service) InitRoutes() {
	gr := s.r.Methods("GET").Subrouter()
	gr.HandleFunc("/hunts", s.GetHunts)
	gr.HandleFunc("/hunts/{hunt_title}", s.GetHunt)
	gr.HandleFunc("/hunts/{hunt_title}/clue", s.NewClue)
	gr.HandleFunc("/photos", s.GetPhotos)
	gr.HandleFunc("/reset-magic", s.GetReset)

	pr := s.r.Methods("POST").Subrouter()
	pr.HandleFunc("/photo", s.PostPhoto)
	pr.HandleFunc("/hunt", s.PostHunt)
	pr.HandleFunc("/hunts/{hunt_title}/clue", s.PostClue)
}

// GET /reset
//
// REMOVES ALL HUNTS!
func (s *Service) GetReset(w http.ResponseWriter, r *http.Request) {
	s.t.Reset()
}

// GET /hunts
//
// Get all the treasure hunts
func (s *Service) GetHunts(w http.ResponseWriter, r *http.Request) {
	hunts := []Hunt{}
	for _, h := range s.t.Hunts() {
		hunts = append(hunts, *h)
	}

	// send JSON payload to clients
	if strings.HasPrefix(r.Header.Get("Content-Type"), "application/json") {
		body, err := json.Marshal(hunts)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		fmt.Fprintf(w, "%v", string(body))
		return
	}

	t, err := template.ParseFiles("site/index.html")
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	if err := t.Execute(w, hunts); err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
}

// GET /hunt/{title}
//
// Get single treasure hunt
func (s *Service) GetHunt(w http.ResponseWriter, r *http.Request) {
	t, err := template.ParseFiles("site/hunt_1.html")
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	// get hunt using title from URL path
	h, ok := s.t.Hunts()[mux.Vars(r)["hunt_title"]]
	if !ok {
		http.Error(w, "Hunt does not exist", http.StatusNotFound)
		return
	}

	// send JSON payload to clients
	if strings.HasPrefix(r.Header.Get("Content-Type"), "application/json") {
		body, err := json.Marshal(h)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		fmt.Fprintf(w, "%v", string(body))
		return
	}

	if err := t.Execute(w, h); err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
}

// POST /hunt
//
// Create a new treasure hunt
func (s *Service) PostHunt(w http.ResponseWriter, r *http.Request) {
	// read form fields
	if err := r.ParseForm(); err != nil {
		http.Error(w, "Could not parse form", http.StatusInternalServerError)
		return
	}

	// try to create a new hunt
	h := Hunt{}
	if h.Name = strings.TrimSpace(r.PostFormValue("title")); len(h.Name) == 0 {
		http.Error(w, "Invalid hunt name", http.StatusBadRequest)
		return
	}

	if h.Description = r.PostFormValue("description"); len(h.Description) == 0 {
		http.Error(w, "Invalid hunt description", http.StatusBadRequest)
		return
	}

	// add the hunt
	if err := s.t.NewHunt(h); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	http.Redirect(w, r, "/hunts", http.StatusFound)
}

// GET /hunts/{hunt_title}/clue
func (s *Service) NewClue(w http.ResponseWriter, r *http.Request) {
	t, err := template.ParseFiles("site/new.html")
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	// build payload
	h, ok := s.t.Hunts()[mux.Vars(r)["hunt_title"]]
	if !ok {
		http.Error(w, "Hunt does not exist", http.StatusNotFound)
		return
	}

	type Payload struct {
		HuntTitle string
		Photos    map[string]Photo
	}

	p := Payload{
		HuntTitle: h.Name,
		Photos:    s.t.Photos(),
	}

	if err := t.Execute(w, p); err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
}

// POST /hunts/{hunt_title}/clue
func (s *Service) PostClue(w http.ResponseWriter, r *http.Request) {
	// read form fields
	if err := r.ParseForm(); err != nil {
		http.Error(w, "Could not parse form", http.StatusInternalServerError)
		return
	}

	// try to create a new clue
	u, err := uuid.NewV4()
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	c := Clue{UUID: u.String()}

	c.Name = r.PostFormValue("name")
	c.Fact = r.PostFormValue("fact")
	if c.Description = r.PostFormValue("description"); len(c.Description) == 0 {
		http.Error(w, "A clue must have a description", http.StatusBadRequest)
		return
	}

	// do we have the specified photo available?
	var ok bool
	if c.Photo, ok = s.t.Photos()[r.PostFormValue("photo_uuid")]; !ok {
		msg := fmt.Sprintf("No photo with id %v exists", r.PostFormValue("photo_uuid"))
		http.Error(w, msg, http.StatusBadRequest)
		return
	}

	if err := s.t.AddClue(mux.Vars(r)["hunt_title"], c); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	// redirect back to hunt
	http.Redirect(w, r, "/hunts/"+mux.Vars(r)["hunt_title"], http.StatusFound)
}

// GET /photos
func (s *Service) GetPhotos(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintf(w, "%+v", s.t.Photos())
}

// POST /photo
func (s *Service) PostPhoto(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	//get the multipart reader for the request.
	reader, err := r.MultipartReader()

	if err != nil {
		out := fmt.Sprintf(`{"msg": %v}`, err.Error())
		http.Error(w, out, http.StatusInternalServerError)
		return
	}

	//copy each part to destination.
	for {
		part, err := reader.NextPart()
		if err == io.EOF {
			break
		}

		//if part.FileName() is empty, skip this iteration.
		if part.FileName() == "" {
			continue
		}
		// generate a uuid for file
		u, err := uuid.NewV4()
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		// TODO should inspect MIME type here.
		name := u.String() + ".jpg"
		pth := "/tmp/teacher-photos/" + name
		dst, err := os.Create(pth)
		defer dst.Close()

		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		if _, err := io.Copy(dst, part); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		// generate a new Photo type and add to service
		photo := Photo{
			UUID: u.String(),
			Path: pth,
			URL:  "http://188.226.156.181/photos/" + name,
		}

		// add the photo to the service
		s.t.AddPhoto(photo)
	}
	//display success message.
	w.WriteHeader(http.StatusNoContent)
}
