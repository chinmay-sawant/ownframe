package inputlab

import "context"

// focusField clicks the center of id so it takes focus.
func (a *App) focusField(ctx context.Context, id string) bool {
	for _, box := range a.page.Boxes() {
		if box.ID != id || box.W <= 0 || box.H <= 0 {
			continue
		}
		if err := a.page.Click(ctx, box.X+box.W/2, box.Y+box.H/2); err != nil {
			return false
		}
		return a.page.Edit().FocusedField() == id
	}
	return false
}

// focusLast focuses the last clipboard field for button chords.
func (a *App) focusLast(ctx context.Context) bool {
	if a.page.Edit().FocusedField() != "" {
		return true
	}
	return a.focusField(ctx, a.lastClip)
}

// focusNote refocuses the edit note after a button blurred it.
func (a *App) focusNote(ctx context.Context) bool {
	return a.focusField(ctx, "e-note")
}

// scrollEnd scrolls past any content length.
const scrollEnd = 1 << 30

// scrollTo moves the page; the window consumes ScrollTo.
func (a *App) scrollTo(x, y int) {
	a.page.ScrollTo(x, y)
}
