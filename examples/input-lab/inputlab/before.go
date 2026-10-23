package inputlab

import (
	"context"
	"github.com/chinmay-sawant/ownframe"
)

// lockedMsg is the alert shown when the locked name takes an edit.
const lockedMsg = "name is locked"

// isLocked reports whether the locked name still holds its preset.
func (a *App) isLocked() bool {
	return a.page.Form().FormValue("l-name") == "locked"
}

// onBeforeEdit snapshots undo stacks and clears old sign-in messages.
// The locked field is never vetoed with an error: an error from here
// would close the whole window, so the edit runs and onChange reverts it.
func (a *App) onBeforeEdit(_ context.Context, box ownframe.Box) error {
	switch box.ID {
	case "e-note", "":
		a.pushNote()
	case "c-left", "c-right":
		a.pushClip()
	case "si-email", "si-password":
		a.pushLogin()
		a.view.LoginOk = ""
		a.view.LoginEr = ""
	}
	return nil
}

// onSubmit signs in on Enter when both fields carry text.
func (a *App) onSubmit(_ context.Context) error {
	if !a.ready() {
		return nil
	}
	a.signIn()
	a.page.SetData(&a.view)
	return nil
}
