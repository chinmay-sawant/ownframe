package telegram

// SetInsets records the system bar insets in CSS pixels. The activity calls
// it from the UI thread; the next tick pads the page. Values clamp to
// 0..2048 so a rogue host cannot corrupt the pinned layout.
func (a *App) SetInsets(top, bottom int) {
	a.insetTop.Store(int64(clampInset(top)))
	a.insetBot.Store(int64(clampInset(bottom)))
}

// maxInset caps one system bar inset in CSS px, so a rogue host cannot
// corrupt the pinned layout.
const maxInset = 2048

func clampInset(v int) int {
	if v < 0 {
		return 0
	}
	if v > maxInset {
		return maxInset
	}
	return v
}

// applyInsets pads the page for the system bars and keeps the newest
// message above the keyboard when the bottom inset grows.
func (a *App) applyInsets() {
	top, bottom := int(a.insetTop.Load()), int(a.insetBot.Load())
	if top == a.view.InsetTop && bottom == a.view.InsetBottom {
		return
	}

	if bottom > a.view.InsetBottom && a.view.Active != "" {
		a.page.ScrollTo(0, scrollBottom)
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
