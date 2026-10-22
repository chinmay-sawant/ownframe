package clipboard

import "context"

// CopyText mirrors the copy chord: the focused field when one is focused,
// and the left field otherwise.
func (a *App) CopyText(ctx context.Context) (string, bool, error) {
	return a.page.Copy(ctx)
}

// CutText mirrors the cut chord. The handler clears the last field.
func (a *App) CutText(ctx context.Context) (string, bool, error) {
	return a.page.Cut(ctx)
}

// PasteText mirrors the paste chord and inserts text in the focused field.
func (a *App) PasteText(ctx context.Context, text string) error {
	return a.page.Paste(ctx, text)
}

// SelectAll mirrors the select-all chord.
func (a *App) SelectAll(ctx context.Context) error {
	return a.page.SelectAll(ctx)
}

// Undo restores both fields from the last snapshot.
func (a *App) Undo(ctx context.Context) error {
	return a.page.Undo(ctx)
}

// Redo reapplies the last undone snapshot.
func (a *App) Redo(ctx context.Context) error {
	return a.page.Redo(ctx)
}

// Type appends text to the focused field.
func (a *App) Type(ctx context.Context, text string) error {
	return a.page.Type(ctx, text)
}

// FormValue returns the current value of id, or "" when id is absent.
func (a *App) FormValue(id string) string {
	return a.page.FormValue(id)
}

// focusLast returns true when a field is focused. When none is, it clicks
// the last field so a button acts where the user was typing.
func (a *App) focusLast(ctx context.Context) bool {
	if a.page.FocusedField() != "" {
		return true
	}

	for _, box := range a.page.Boxes() {
		if box.ID != a.last || box.W <= 0 || box.H <= 0 {
			continue
		}

		if err := a.page.Click(ctx, box.X+box.W/2, box.Y+box.H/2); err != nil {
			return false
		}

		return a.page.FocusedField() == a.last
	}

	return false
}
