package model

import "time"

type Pet struct {
	ID, Name, Species, Breed, BirthDate, Notes string
}

type Record struct {
	ID, PetID, Category, Date, Note string
	Scheduled, Complete             bool
}

type Backup struct {
	Version int      `json:"version"`
	Pets    []Pet    `json:"pets"`
	Records []Record `json:"records"`
}

func Today() string { return time.Now().Format("2006-01-02") }
