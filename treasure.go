package main

import (
	"fmt"
	"sync"
)

type TreasureHunts struct {
	hunts  map[string]*Hunt
	photos map[string]Photo
	*sync.Mutex
}

func (t *TreasureHunts) Reset() {
	t.Lock()
	defer t.Unlock()
	t.hunts = map[string]*Hunt{}
	t.photos = map[string]Photo{}
}

// create a new hunt
func (t *TreasureHunts) NewHunt(h Hunt) error {
	t.Lock()
	defer t.Unlock()
	if _, ok := t.hunts[h.Name]; ok {
		// hunt already exists
		return fmt.Errorf("Treasure hunt with name: %v already exists.", h.Name)
	}
	t.hunts[h.Name] = &h
	return nil
}

// get all the hunts
func (t *TreasureHunts) Hunts() map[string]*Hunt {
	t.Lock()
	defer t.Unlock()
	return t.hunts
}

func (t *TreasureHunts) AddClue(huntTitle string, c Clue) error {
	t.Lock()
	defer t.Unlock()
	// does the hunt exist?
	h, ok := t.hunts[huntTitle]
	if !ok {
		// hunt does not exist
		return fmt.Errorf("Treasure hunt with name: %v does not exist.", huntTitle)
	}

	// add the clue to the clue list in the hunt
	h.Clues = append(h.Clues, c)
	return nil
}

func (t *TreasureHunts) AddPhoto(p Photo) {
	t.Lock()
	defer t.Unlock()
	t.photos[p.UUID] = p
}

// get all the photos
func (t *TreasureHunts) Photos() map[string]Photo {
	t.Lock()
	defer t.Unlock()
	return t.photos
}

type Hunt struct {
	Name        string `json:"title"`
	Description string `json:"description"`
	Clues       []Clue `json:"clues"`
}

type Clue struct {
	UUID        string `json:"uuid"`
	Photo       Photo  `json:"photo"`
	Name        string `json:"name"`
	Description string `json:"description"`
}

type Photo struct {
	UUID string `json:"uuid"`
	URL  string `json:"url"`
	Path string `json:"-"`
}
