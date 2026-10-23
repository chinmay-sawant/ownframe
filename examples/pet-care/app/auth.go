package app

func (a *App) RequestBack() bool {
	if !a.canBack.Load() {
		return false
	}
	a.back.Store(true)
	return true
}

func (a *App) login() {
	// Demo pair only. DO NOT COPY into production; verify server side
	// over TLS with a slow hash, never a plaintext == comparison.
	if a.page.Form().FormValue("email") == "demo@petcare.local" && a.page.Form().FormValue("password") == "pawprints" {
		a.view.SignedIn = true
		a.view.Screen = "dashboard"
		a.view.Error = ""
		a.view.Status = "Welcome back."
		return
	}
	a.view.Error = "Those demo details do not match. Try the prefilled values."
}
