package app

import (
	"context"
	"github.com/chinmay-sawant/ownframe/examples/pet-care/model"
	"strings"
)

// recentLimit is how many recent records the dashboard keeps.
const recentLimit = 5

func (a *App) reload(ctx context.Context) error {
	if a.db == nil {
		a.view.StorageWarning = "Local storage could not open. Changes stay temporary for this session."
		return nil
	}
	pets, err := a.db.Pets(ctx)
	if err != nil {
		return err
	}
	records, err := a.db.Records(ctx, "")
	if err != nil {
		return err
	}
	a.view.AllPets = pets
	a.view.Pets = pets
	if a.view.Screen == "pets" || a.view.Screen == "history" {
		a.view.Search = a.page.FormValue("search")
		if a.view.Search != "" {
			filtered := make([]model.Pet, 0, len(pets))
			for _, p := range pets {
				if strings.Contains(strings.ToLower(p.Name+" "+p.Species), strings.ToLower(a.view.Search)) {
					filtered = append(filtered, p)
				}
			}
			a.view.Pets = filtered
		}
	}
	rows := recordRows(records, pets)
	a.view.Recent = rows
	if len(a.view.Recent) > recentLimit {
		a.view.Recent = a.view.Recent[:recentLimit]
	}
	a.view.Records = visibleRecords(rows, a.view.Screen, a.view.Selected.ID, a.view.Search)
	a.view.HasSelected = false
	for _, p := range pets {
		if p.ID == a.view.Selected.ID {
			a.view.Selected = p
			a.view.HasSelected = true
		}
	}
	if a.view.Selected.ID != "" && !a.view.HasSelected {
		a.view.Selected = model.Pet{}
	}
	return nil
}

func (a *App) redraw(ctx context.Context) error {
	if err := a.reload(ctx); err != nil {
		a.view.Error = err.Error()
	}
	a.page.SetData(a.view)
	a.canBack.Store(a.view.SignedIn && a.view.Screen != "dashboard")
	return nil
}
