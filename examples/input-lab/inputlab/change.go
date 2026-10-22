package inputlab

import (
	"context"
	"fmt"

	"github.com/chinmay-sawant/ownframe"
)

// onChange mirrors the controls demo: bound ids report on their own
// lines, everything else refreshes the global status line.
func (a *App) onChange(_ context.Context, box ownframe.Box) error {
	switch {
	case len(box.ID) > 2 && box.ID[:2] == "b-":
		a.view.BStatus = "changed " + box.ID
		a.page.SetData(&a.view)
		return nil
	case len(box.ID) > 2 && box.ID[:2] == "l-":
		if box.ID == "l-name" && !a.isLocked() {
			a.page.Form().SetFormValue("l-name", "locked")
			a.view.LName = "locked"
			a.view.LStatus = lockedMsg
			a.page.SetData(&a.view)
			return nil
		}
		a.view.LStatus = fmt.Sprintf("changed %s value=%s",
			box.ID, a.valueOf(box.ID))
		a.page.SetData(&a.view)
		return nil
	}
	if box.ID == "si-email" || box.ID == "si-password" {
		a.syncReady()
	}
	a.refresh()
	return nil
}

// valueOf reads the control that just changed.
func (a *App) valueOf(id string) string {
	if id == "l-agree" || id == "l-free" || id == "l-pro" {
		return fmt.Sprintf("%t", a.page.Form().FormChecked(id))
	}
	return a.page.Form().FormValue(id)
}
