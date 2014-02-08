package main

import (
	"fmt"
	"github.com/gorilla/mux"
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

	pr := s.r.Methods("POST").Subrouter()
	pr.HandleFunc("/photo", s.PostPhoto)
	pr.HandleFunc("/hunt", s.PostHunt)
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
	// get titls from URL path
	title := mux.Vars(r)["hunt_title"]
	h, ok := s.t.Hunts()[title]
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
	}

	if h.Description = r.PostFormValue("description"); len(h.Description) == 0 {
		http.Error(w, "Invalid hunt description", http.StatusBadRequest)
	}

	// add the hunt
	if err := s.t.NewHunt(h); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
	}
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
		dst, err := os.Create("/tmp/teacher-photos/" + part.FileName())
		defer dst.Close()

		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		if _, err := io.Copy(dst, part); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
	}
	//display success message.
	w.WriteHeader(http.StatusNoContent)
}
