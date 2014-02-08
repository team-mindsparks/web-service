package main

import (
	"fmt"
	"github.com/gorilla/mux"
	"github.com/nu7hatch/gouuid"
	"io"
	"net/http"
	"os"
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
	gr.HandleFunc("/photos", s.GetPhotos)

	pr := s.r.Methods("POST").Subrouter()
	pr.HandleFunc("/photo", s.PostPhoto)
	pr.HandleFunc("/hunt", s.PostHunt)
	pr.HandleFunc("/hunts/{hunt_title}/clue", s.PostClue)
}

// GET /hunts
//
// Get all the treasure hunts
func (s *Service) GetHunts(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintln(w, fmt.Sprintf("%+v", s.t.Hunts()))
}

// GET /hunt/{title}
//
// Get single treasure hunt
func (s *Service) GetHunt(w http.ResponseWriter, r *http.Request) {
	// get hunt using title from URL path
	h, ok := s.t.Hunts()[mux.Vars(r)["hunt_title"]]
	if !ok {
		http.Error(w, "Hunt does not exist", http.StatusNotFound)
		return
	}
	fmt.Fprintln(w, fmt.Sprintf("+%v", h))
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
	if h.Name = r.PostFormValue("title"); len(h.Name) == 0 {
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
