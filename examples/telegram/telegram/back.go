package telegram

// RequestBack accepts a back press on the phone. It reports false when the
// app is already on a list, so the activity can close instead.
func (a *App) RequestBack() bool {
	if a.onList.Load() {
		return false
	}

	a.mu.Lock()
	a.back = true
	a.mu.Unlock()

	return true
}

// applyBack leaves the open chat on the next tick.
func (a *App) applyBack() {
	a.mu.Lock()
	back := a.back
	a.back = false
	a.mu.Unlock()

	if !back {
		return
	}

	a.view.Active = ""
	a.view.Status = ""
	a.page.ScrollTo(0, 0)
	a.mark()
}
