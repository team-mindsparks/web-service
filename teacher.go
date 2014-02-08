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
}

func (s *Service) InitRoutes() {
	postRouter := s.r.Methods("POST").Subrouter()
	postRouter.HandleFunc("/photo", s.PostPhoto)
}

func (s *Service) PostPhoto(w http.ResponseWriter, r *http.Request) {
	//get the multipart reader for the request.
	reader, err := r.MultipartReader()

	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
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
	fmt.Fprintln(w, "Sorted!")
}
