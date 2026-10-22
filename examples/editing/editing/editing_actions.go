package editing

import "context"

// SelectAll selects the whole focused note.
func (a *App) SelectAll(ctx context.Context) error {
	return a.page.SelectAll(ctx)
}

// selectAllButton refocuses the note before selecting it. The click that
// opens the button blurs the form first.
func (a *App) selectAllButton(ctx context.Context) error {
	if !a.focusNote(ctx) {
		a.view.Status = "select-all needs a focused field"
		a.page.SetData(a.view)

		return nil
	}

	return a.page.SelectAll(ctx)
}

// focusNote clicks the note field so it takes focus.
func (a *App) focusNote(ctx context.Context) bool {
	for _, box := range a.page.Boxes() {
		if box.ID != "note" || box.W <= 0 || box.H <= 0 {
			continue
		}

		if err := a.page.Click(ctx, box.X+box.W/2, box.Y+box.H/2); err != nil {
			return false
		}

		return a.page.FocusedField() == "note"
	}

	return false
}

// Undo restores the previous note.
func (a *App) Undo(ctx context.Context) error {
	return a.page.Undo(ctx)
}

// Redo restores the note undone last.
func (a *App) Redo(ctx context.Context) error {
	return a.page.Redo(ctx)
}

// FormValue returns the stored value of a control id.
func (a *App) FormValue(id string) string {
	return a.page.FormValue(id)
}

// FormSelected reports whether id is focused with its whole value selected.
func (a *App) FormSelected(id string) bool {
	return a.page.FormSelected(id)
}

// View returns the stamp and status data the template prints.
func (a *App) View() View {
	return a.view
}
