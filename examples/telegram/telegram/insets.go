package telegram

// SetInsets records the system bar insets in CSS pixels. The activity calls
// it from the UI thread; the next tick pads the page.
func (a *App) SetInsets(top, bottom int) {
	a.insetTop.Store(int64(top))
	a.insetBot.Store(int64(bottom))
}

// applyInsets pads the page for the system bars and keeps the newest
// message above the keyboard when the bottom inset grows.
func (a *App) applyInsets() {
	top, bottom := int(a.insetTop.Load()), int(a.insetBot.Load())
	if top == a.view.InsetTop && bottom == a.view.InsetBottom {
		return
	}

	if bottom > a.view.InsetBottom && a.view.Active != "" {
		a.page.ScrollTo(0, 1<<20)
	} else if bottom < a.view.InsetBottom {
		// The keyboard closed. The pinned bars were laid out against the
		// old, taller document, so the window's clamp cannot pull the
		// thread back. Mark a scroll to the thread's real end after the
		// redraw, where the boxes are fresh.
		a.pullThread = true
	}

	a.view.InsetTop, a.view.InsetBottom = top, bottom
	a.view.Phone = top > 0 || bottom > 0
	a.mark()
}
