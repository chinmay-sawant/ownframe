package bindhooks

import (
	"context"
	"errors"
	"fmt"

	"github.com/chinmay-sawant/ownframe"
)

// errLocked aborts an edit that starts from the locked name.
var errLocked = errors.New("bindhooks: name is locked")

// onBeforeEdit vetoes an edit of name while its stored value is "locked".
// Every other control edits normally.
func (a *App) onBeforeEdit(_ context.Context, box ownframe.Box) error {
	if box.ID == "name" && a.page.Form().FormValue("name") == "locked" {
		return errLocked
	}

	return nil
}

// onChange records the new value after the library wrote it into View.
func (a *App) onChange(_ context.Context, box ownframe.Box) error {
	a.view.Status = fmt.Sprintf("changed %s value=%s", box.ID, a.valueOf(box.ID))
	a.page.SetData(&a.view)

	return nil
}

// valueOf reads the control that just changed. Checkboxes and radios report
// FormChecked; the text controls report FormValue.
func (a *App) valueOf(id string) string {
	switch id {
	case "agree", "free", "pro":
		return fmt.Sprintf("%t", a.page.Form().FormChecked(id))
	default:
		return a.page.Form().FormValue(id)
	}
}
