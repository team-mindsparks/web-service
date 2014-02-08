package main

import (
	"fmt"
	"sync"
	"time"
)

type TreasureHunts struct {
	hunts map[string]Hunt
	*sync.Mutex
}

func (t *TreasureHunts) NewHunt(h Hunt) error {
	t.Lock()
	defer t.Unlock()
	if _, ok := t.hunts[h.Name]; ok {
		// hunt already exists
		return fmt.Errorf("Treasure hunt with name: %v already exists.", h.Name)
	}
	t.hunts[h.Name] = h
	return nil
}

func (t *TreasureHunts) Hunts() []Hunt {
	t.Lock()
	defer t.Unlock()
	hunts := []Hunt{}
	for _, v := range t.hunts {
		hunts = append(hunts, v)
	}
	return hunts
}

type Hunt struct {
	Name        string `json: "title"`
	Description string `json: "description"`
	Clues       []Clue `json: "clues"`
}

type Clue struct {
	Id          int     `Id`
	Photo       []Photo `json: "photos"`
	Name        string  `json: "name"`
	Description string  `json: "description"`
}

type Photo struct {
	Id          int       `json:"id"`
	URL         string    `json:"url"`
	Uploaded    time.Time `json: "uploaded"`
	Path        string
	Fingerprint []byte `json:"fingerprint"`
}
