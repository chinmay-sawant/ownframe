package app

import (
	"context"
	"github.com/chinmay-sawant/ownframe/examples/pet-care/model"
	"strings"
	"time"
)

func (a *App) addRecord(ctx context.Context) error {
	petID := a.page.Form().FormValue("record-pet")
	if petID == "" {
		petID = a.view.Selected.ID
	}
	if petID == "" && len(a.view.Pets) > 0 {
		petID = a.view.Pets[0].ID
	}
	category := strings.TrimSpace(a.page.Form().FormValue("record-category"))
	date := a.page.Form().FormValue("record-date")
	if petID == "" || category == "" {
		a.view.Error = "Choose a pet and enter a care type."
		return a.redraw(ctx)
	}
	if _, err := time.Parse("2006-01-02", date); err != nil {
		a.view.Error = "Enter a valid care date."
		return a.redraw(ctx)
	}
	r := model.Record{ID: id(), PetID: petID, Category: category, Date: date, Note: a.page.Form().FormValue("record-note"), Scheduled: a.page.Form().FormChecked("record-scheduled")}
	if a.db == nil {
		a.view.Error = "This session is temporary, so this record cannot be saved."
		return a.redraw(ctx)
	}
	if err := a.db.SaveRecord(ctx, r); err != nil {
		a.view.Error = err.Error()
		return a.redraw(ctx)
	}
	a.view.Error = ""
	a.view.Status = "Care record saved."
	a.view.Screen = "history"
	return a.redraw(ctx)
}

func (a *App) toggleRecord(ctx context.Context, id string) error {
	if a.db == nil {
		a.view.Error = "Local storage is unavailable."
		return a.redraw(ctx)
	}
	for _, r := range a.view.Records {
		if r.ID == id {
			r.Complete = !r.Complete
			if err := a.db.SaveRecord(ctx, r.Record); err != nil {
				a.view.Error = err.Error()
			}
			break
		}
	}
	return a.redraw(ctx)
}

func (a *App) removeRecord(ctx context.Context, id string) error {
	if a.db == nil {
		a.view.Error = "Local storage is unavailable."
		return a.redraw(ctx)
	}
	if err := a.db.DeleteRecord(ctx, id); err != nil {
		a.view.Error = err.Error()
	}
	a.view.ConfirmRecord = ""
	return a.redraw(ctx)
}
