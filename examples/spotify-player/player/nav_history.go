package player

// navigate records a screen change and drops any forward branch.
func (a *App) navigate(screen string) {
	if screen == a.view.Nav {
		return
	}
	a.navHistory = append(a.navHistory[:a.navAt+1], screen)
	a.navAt = len(a.navHistory) - 1
	a.view.Nav = screen
	a.updateNavButtons()
}

// moveNav restores the previous or next screen when one exists.
func (a *App) moveNav(delta int) {
	at := a.navAt + delta
	if at < 0 || at >= len(a.navHistory) {
		return
	}
	a.navAt = at
	a.view.Nav = a.navHistory[at]
	a.updateNavButtons()
}

// updateNavButtons keeps the arrows aligned with the current history entry.
func (a *App) updateNavButtons() {
	a.view.CanBack = a.navAt > 0
	a.view.CanForward = a.navAt+1 < len(a.navHistory)
}
