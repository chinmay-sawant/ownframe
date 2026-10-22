package app

import (
	"crypto/rand"
	"encoding/hex"
	"fmt"
)

func (a *App) toggleTheme() {
	a.view.Dark = !a.view.Dark
	a.dark.Store(a.view.Dark)
	if a.view.Dark {
		_ = a.page.SetTheme(darkCSS)
	} else {
		_ = a.page.SetTheme("")
	}
}

func (a *App) navigate(screen string) {
	a.view.Screen = screen
	if screen == "add-pet" {
		a.view.PetName = ""
		a.view.PetSpecies = ""
		a.view.PetBreed = ""
		a.view.PetBirth = ""
		a.view.PetNotes = ""
		a.view.EditPetID = ""
	}
}

func (a *App) editPet() {
	p := a.view.Selected
	a.view.PetName = p.Name
	a.view.PetSpecies = p.Species
	a.view.PetBreed = p.Breed
	a.view.PetBirth = p.BirthDate
	a.view.PetNotes = p.Notes
	a.view.EditPetID = p.ID
	a.view.Screen = "add-pet"
}

func (a *App) selectPet(id string) {
	for _, p := range a.view.AllPets {
		if p.ID == id {
			a.view.Selected = p
			a.view.HasSelected = true
			a.view.Screen = "profile"
			return
		}
	}
	a.view.Error = fmt.Sprintf("Pet %s was not found.", id)
}

func id() string {
	var raw [16]byte
	if _, err := rand.Read(raw[:]); err != nil {
		panic(err)
	}
	return hex.EncodeToString(raw[:])
}
