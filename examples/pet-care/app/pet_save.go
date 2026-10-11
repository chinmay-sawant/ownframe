package app

import (
	"context"
	"github.com/chinmay-sawant/ownframe/examples/pet-care/model"
	"strings"
	"time"
)

func (a *App) addPet(ctx context.Context) error {
	petID := a.view.EditPetID
	if petID == "" {
		petID = id()
	}
	p := model.Pet{ID: petID, Name: strings.TrimSpace(a.page.FormValue("pet-name")), Species: strings.TrimSpace(a.page.FormValue("pet-species")), Breed: a.page.FormValue("pet-breed"), BirthDate: a.page.FormValue("pet-birth"), Notes: a.page.FormValue("pet-notes")}
	if p.Name == "" || p.Species == "" {
		a.view.Error = "Add a name and species before saving."
		return a.redraw(ctx)
	}
	if p.BirthDate != "" {
		if _, err := time.Parse("2006-01-02", p.BirthDate); err != nil || p.BirthDate > model.Today() {
			a.view.Error = "Enter a valid birth date that is not in the future."
			return a.redraw(ctx)
		}
	}
	if err := a.db.SavePet(ctx, p); err != nil {
		a.view.Error = err.Error()
		return a.redraw(ctx)
	}
	a.view.Selected = p
	a.view.HasSelected = true
	a.view.EditPetID = ""
	a.view.Screen = "profile"
	a.view.Error = ""
	a.view.Status = "Pet saved."
	return a.redraw(ctx)
}
