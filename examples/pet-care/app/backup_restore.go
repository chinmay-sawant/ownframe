package app

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/chinmay-sawant/ownframe/examples/pet-care/model"
	"github.com/chinmay-sawant/ownframe/examples/pet-care/store"
	"strings"
	"time"
)

func (a *App) restoreBackup(ctx context.Context) error {
	var b model.Backup
	if err := json.Unmarshal([]byte(a.view.BackupJSON), &b); err != nil {
		a.view.Error = "That backup could not be read."
		return a.redraw(ctx)
	}
	if err := validate(b); err != nil {
		a.view.Error = err.Error()
		return a.redraw(ctx)
	}
	if err := a.db.Replace(ctx, b); err != nil {
		a.view.Error = "Restore failed. Existing data is unchanged."
		return a.redraw(ctx)
	}
	a.view.ConfirmRestore = false
	a.view.Status = "Backup restored."
	a.view.Screen = "dashboard"
	return a.redraw(ctx)
}

func validate(b model.Backup) error {
	if err := store.ValidateBackup(b); err != nil {
		return err
	}
	if len(b.Pets) > 1000 || len(b.Records) > 10000 {
		return fmt.Errorf("backup is too large")
	}
	for _, p := range b.Pets {
		if strings.TrimSpace(p.Name) == "" || strings.TrimSpace(p.Species) == "" {
			return fmt.Errorf("pet names and species cannot be blank")
		}
		if p.BirthDate != "" {
			d, err := time.Parse("2006-01-02", p.BirthDate)
			if err != nil || d.After(time.Now()) {
				return fmt.Errorf("backup has an invalid birth date")
			}
		}
	}
	return nil
}
