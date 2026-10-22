package login

// ready reports whether both fields carry text, so the sign-in button can
// turn blue and take a click only after the user filled them.
func (a *App) ready() bool {
	return a.view.Email != "" && a.view.Password.Reveal() != ""
}

// refresh syncs View.Ready with the fields. The redraw that follows every
// handler reads the new value from the template data.
func (a *App) refresh() {
	ready := a.ready()
	if a.view.Ready == ready {
		return
	}

	a.view.Ready = ready
	a.page.SetData(&a.view)
}
