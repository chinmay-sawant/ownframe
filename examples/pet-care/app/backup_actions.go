package app

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/chinmay-sawant/ownframe/examples/pet-care/model"
	"strings"
)

func (a *App) exportBackup(ctx context.Context) error {
	b, err := a.db.Backup(ctx)
	if err != nil {
		a.view.Error = err.Error()
		return a.redraw(ctx)
	}
	raw, err := json.MarshalIndent(b, "", "  ")
	if err != nil {
		a.view.Error = err.Error()
		return a.redraw(ctx)
	}
	a.view.BackupJSON = string(raw)
	a.view.Error = ""
	a.view.Status = "Backup ready. Copy the JSON before leaving this screen."
	a.requestExport(string(raw))
	a.requestPicker("export")
	return a.redraw(ctx)
}

func (a *App) checkBackup(ctx context.Context) error {
	raw := strings.TrimSpace(a.page.FormValue("backup-json"))
	if len(raw) == 0 || len(raw) > 1<<20 {
		a.view.Error = "Paste a backup smaller than 1 MB."
		return a.redraw(ctx)
	}
	var b model.Backup
	if err := json.Unmarshal([]byte(raw), &b); err != nil {
		a.view.Error = "That backup is not valid JSON."
		return a.redraw(ctx)
	}
	if err := validate(b); err != nil {
		a.view.Error = err.Error()
		return a.redraw(ctx)
	}
	a.view.BackupJSON = raw
	a.view.ConfirmRestore = true
	a.view.RestoreMessage = fmt.Sprintf("Replace local data with %d pets and %d care records?", len(b.Pets), len(b.Records))
	return a.redraw(ctx)
}
