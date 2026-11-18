package telegram

// SetWebMode switches the thread chrome to browser-fixed positioning for -web DOM mode.
func (a *App) SetWebMode(on bool) {
	a.web = on
	a.view.Web = on
	if on {
		a.threadLimit = webThreadPage
	} else {
		a.threadLimit = 0
	}
	a.rebuild()
	a.page.SetData(&a.view)
}
