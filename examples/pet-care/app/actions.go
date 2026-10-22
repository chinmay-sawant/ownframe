package app

import (
	"context"
	"github.com/chinmay-sawant/ownframe"
	"strings"
)

func (a *App) click(ctx context.Context, b ownframe.Box) error {
	action := b.Action
	switch {
	case action == "login":
		a.login()
	case action == "signout":
		a.view.SignedIn = false
		a.view.Screen = "login"
		a.view.Status = ""
	case strings.HasPrefix(action, "nav:"):
		a.navigate(strings.TrimPrefix(action, "nav:"))
	case action == "edit-pet":
		a.editPet()
	case strings.HasPrefix(action, "pet:"):
		a.selectPet(strings.TrimPrefix(action, "pet:"))
	case action == "theme":
		a.toggleTheme()
	case action == "backup-export":
		return a.exportBackup(ctx)
	case action == "backup-restore":
		return a.checkBackup(ctx)
	case action == "backup-picker":
		a.requestPicker("restore")
		a.view.Status = "Android file picker requested. On desktop, paste the JSON backup here."
	case action == "restore-confirm":
		return a.restoreBackup(ctx)
	case action == "restore-cancel":
		a.view.ConfirmRestore = false
	case action == "cancel-delete":
		a.view.ConfirmPet = ""
	case action == "cancel-remove":
		a.view.ConfirmRecord = ""
	case action == "add-pet":
		return a.addPet(ctx)
	case action == "add-record":
		return a.addRecord(ctx)
	case strings.HasPrefix(action, "done:"):
		return a.toggleRecord(ctx, strings.TrimPrefix(action, "done:"))
	case strings.HasPrefix(action, "delete:"):
		a.view.ConfirmPet = strings.TrimPrefix(action, "delete:")
	case action == "confirm-delete":
		return a.deletePet(ctx)
	case strings.HasPrefix(action, "remove-record:"):
		a.view.ConfirmRecord = strings.TrimPrefix(action, "remove-record:")
	case strings.HasPrefix(action, "confirm-remove:"):
		return a.removeRecord(ctx, strings.TrimPrefix(action, "confirm-remove:"))
	}
	return a.redraw(ctx)
}

func (a *App) change(ctx context.Context, _ ownframe.Box) error { return a.redraw(ctx) }
func (a *App) submit(ctx context.Context) error                 { a.login(); return a.redraw(ctx) }
