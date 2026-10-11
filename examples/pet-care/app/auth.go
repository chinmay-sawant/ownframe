package app

import "context"

func (a *App) Back() bool {
	if !a.view.SignedIn {
		return false
	}
	if a.view.Screen == "profile" {
		a.view.Screen = "pets"
	} else {
		a.view.Screen = "dashboard"
	}
	_ = a.redraw(context.Background())
	return true
}

func (a *App) login() {
	if a.page.FormValue("email") == "demo@petcare.local" && a.page.FormValue("password") == "pawprints" {
		a.view.SignedIn = true
		a.view.Screen = "dashboard"
		a.view.Error = ""
		a.view.Status = "Welcome back."
		return
	}
	a.view.Error = "Those demo details do not match. Try the prefilled values."
}
