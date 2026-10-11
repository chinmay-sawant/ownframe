package app

import (
	"context"
	"fmt"
)

func (a *App) deletePet(ctx context.Context) error {
	if a.view.ConfirmPet == "" {
		return nil
	}
	if a.db == nil {
		a.view.Error = "This session is temporary, so there is no saved pet to delete."
		return a.redraw(ctx)
	}
	if err := a.db.DeletePet(ctx, a.view.ConfirmPet); err != nil {
		a.view.Error = fmt.Sprintf("Could not delete pet and its records: %v", err)
	} else {
		a.view.Status = "Pet and its care records deleted."
	}
	a.view.ConfirmPet = ""
	a.view.Selected.ID = ""
	a.view.HasSelected = false
	a.view.Screen = "pets"
	return a.redraw(ctx)
}
