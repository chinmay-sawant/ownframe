package login

import "github.com/chinmay-sawant/ownframe"

// snap is one saved pair of field values. The password stays a
// Secret so a %v or slog dump of undo history cannot leak it.
type snap struct {
	Email    string
	Password ownframe.Secret
}

func (a *App) signIn() {
	status := ""
	errMsg := "Unknown email or password."
	// Demo credentials only. DO NOT COPY into production: no auth
	// boundary here; a real app must verify server side, not locally.
	if a.view.Email == "secret" && a.view.Password.Reveal() == "secret" {
		status = "Signed in."
		errMsg = ""
	}

	if a.view.Status == status && a.view.Error == errMsg {
		return
	}

	a.view.Status = status
	a.view.Error = errMsg
}

// snapshot copies both bound fields.
func (a *App) snapshot() snap {
	return snap{Email: a.view.Email, Password: a.view.Password}
}

// push saves the current fields for undo. A snapshot equal to the last one
// is dropped so no-op edits do not stack. Any redo history is discarded.
func (a *App) push() {
	next := a.snapshot()
	if n := len(a.undo); n == 0 || a.undo[n-1] != next {
		a.undo = append(a.undo, next)
		if len(a.undo) > undoLimit {
			a.undo = a.undo[len(a.undo)-undoLimit:]
		}
	}

	a.redo = nil
}

// restore writes a snapshot into View. The next Redraw pushes it.
func (a *App) restore(s snap) {
	a.view.Email = s.Email
	a.view.Password = s.Password
	a.page.SetFormValue("email", s.Email)
	a.page.SetFormValue("password", s.Password.Reveal())
}
