package controls

import "context"

// Type edits the focused field. A nil type handler does not stop the edit.
func (a *App) Type(ctx context.Context, text string) error {
	return a.page.Type(ctx, text)
}

// Backspace drops the last rune of the focused field.
func (a *App) Backspace(ctx context.Context) error {
	return a.page.Backspace(ctx)
}

// DeleteWord drops the last word of the focused field.
func (a *App) DeleteWord(ctx context.Context) error {
	return a.page.DeleteWord(ctx)
}

// FormValue returns the stored value of a control id.
func (a *App) FormValue(id string) string {
	return a.page.FormValue(id)
}

// FormChecked reports whether a checkbox or a radio is checked.
func (a *App) FormChecked(id string) bool {
	return a.page.FormChecked(id)
}

// FocusedField returns the focused control id, or "" when none is focused.
func (a *App) FocusedField() string {
	return a.page.FocusedField()
}

// View returns the status data the template prints.
func (a *App) View() View {
	return a.view
}
