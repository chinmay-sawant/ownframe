package login

import "github.com/chinmay-sawant/ownframe/internal/secret"

// snap is one saved pair of field values. The password stays a
// secret.Secret so a %v or slog dump of undo history cannot leak it.
type snap struct {
	Email    string
	Password secret.Secret
}

func (a *App) signIn() {
	status := ""
	errMsg := "Unknown email or password."
	// Demo credentials only. DO NOT COPY into production: no auth
	// boundary here; a real app must verify server side, not locally.
	if a.page.FormValue("email") == "secret" && a.page.FormValue("password") == "secret" {
		status = "Signed in."
		errMsg = ""
	}

	if a.view.Status == status && a.view.Error == errMsg {
		return
	}

	a.view.Status = status
	a.view.Error = errMsg
}

// snapshot reads both fields. The library owns their values, not View.
func (a *App) snapshot() snap {
	return snap{
		Email:    a.page.FormValue("email"),
		Password: secret.Secret(a.page.FormValue("password")),
	}
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

// restore writes a snapshot back into the page fields.
func (a *App) restore(s snap) {
	a.page.SetFormValue("email", s.Email)
	a.page.SetFormValue("password", s.Password.Reveal())
}
