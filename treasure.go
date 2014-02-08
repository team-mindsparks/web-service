package main

import (
	"time"
)

type Hunt struct {
	Name        string `json: "title"`
	Description string `json: "description"`
	Clues       []Clue `json: "clues"`
}

type Clue struct {
	Photo       []Photo `json: "photos"`
	Name        string  `json: "name"`
	Description string  `json: "description"`
}

type Photo struct {
	URL         string    `json:"url"`
	Uploaded    time.Time `json: "uploaded"`
	Path        string
	Fingerprint []byte `json:"fingerprint"`
}
