package app

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/chinmay-sawant/ownframe/examples/pet-care/model"
)

func (a *App) requestPicker(kind string) {
	select {
	case a.pickers <- kind:
	default:
	}
}
func (a *App) requestExport(data string) {
	select {
	case a.exports <- []byte(data):
	default:
	}
}

func (a *App) TakePickerRequest() string {
	select {
	case v := <-a.pickers:
		return v
	default:
		return ""
	}
}
func (a *App) TakeBackupData() []byte {
	select {
	case v := <-a.exports:
		return v
	default:
		return nil
	}
}
func (a *App) QueueBackupImport(data []byte) {
	if len(data) > 1<<20 {
		return
	}
	copyData := append([]byte(nil), data...)
	select {
	case a.imports <- copyData:
	default:
	}
}

func (a *App) Tick(ctx context.Context) error {
	changed := false
	select {
	case raw := <-a.imports:
		var b model.Backup
		if err := json.Unmarshal(raw, &b); err != nil {
			a.view.Error = "That backup is not valid JSON."
		} else if err = validate(b); err != nil {
			a.view.Error = err.Error()
		} else {
			a.view.BackupJSON = string(raw)
			a.view.ConfirmRestore = true
			a.view.RestoreMessage = fmt.Sprintf("Replace local data with %d pets and %d care records?", len(b.Pets), len(b.Records))
			a.view.Error = ""
		}
		a.view.Screen = "settings"
		changed = true
	default:
	}
	if a.back.Swap(false) && a.view.SignedIn {
		if a.view.Screen == "profile" {
			a.view.Screen = "pets"
		} else {
			a.view.Screen = "dashboard"
		}
		a.view.Status = ""
		changed = true
	}
	if !changed {
		return nil
	}
	if err := a.redraw(ctx); err != nil {
		return err
	}
	return a.page.Redraw(ctx)
}
