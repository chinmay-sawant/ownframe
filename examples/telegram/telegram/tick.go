package telegram

import "context"

// Tick drains cross-thread requests: a back press, new system bar insets,
// and a captured photo. It redraws when any of them changed the view. The
// window calls it once per frame on the phone.
func (a *App) Tick(ctx context.Context) error {
	a.page.UseFrameDirty()
	if err := a.applyBlur(ctx); err != nil {
		return err
	}
	a.applyBack()
	a.applyInsets()
	a.applyPhoto()

	if a.dirty {
		a.dirty = false
		a.rebuild()
		a.page.SetData(&a.view)

		if err := a.page.Redraw(ctx); err != nil {
			return err
		}

		a.pullThreadEnd()
		a.scrollDirty = false
	}

	return a.settleScroll(ctx)
}

// pullThreadEnd scrolls to the bottom of the rebuilt thread. The inset
// change that marked it cannot rely on the window's clamp: the pinned bars
// were laid out against the old document.
func (a *App) pullThreadEnd() {
	if !a.pullThread {
		return
	}

	a.pullThread = false
	_, viewH := a.page.Size()

	for _, b := range a.Boxes() {
		if b.ID == "thread" {
			offset := int(b.Y+b.H) + a.view.InsetBottom - viewH
			if offset < 0 {
				offset = 0
			}

			a.page.ScrollTo(0, offset)

			return
		}
	}
}

// mark records that the view needs a rebuild on the next tick.
func (a *App) mark() { a.dirty = true }
