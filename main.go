package main

import (
	"github.com/gorilla/mux"
	"log"
	"net/http"
	"sync"
)

func main() {
	// make a new service!
	s := Service{
		r: mux.NewRouter(),
		t: TreasureHunts{
			hunts: map[string]Hunt{},
			Mutex: &sync.Mutex{},
		},
	}

	// init routes
	s.InitRoutes()

	http.Handle("/", s.r)
	log.Fatal(http.ListenAndServe(":8080", http.DefaultServeMux))
}