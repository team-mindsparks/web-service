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
	getRouter := s.r.Methods("GET").Subrouter()
	getRouter.HandleFunc("/hunts", s.GetHunts)

	postRouter := s.r.Methods("POST").Subrouter()
	postRouter.HandleFunc("/photo", s.PostPhoto)
	postRouter.HandleFunc("/hunt", s.PostHunt)
}

// GET /hunts
//
// Get all the treasure hunts
func (s *Service) GetHunts(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintln(w, s.t.Hunts())
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
