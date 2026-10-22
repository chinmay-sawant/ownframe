package inputlab

import "context"

// Type edits the focused field.
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

// Paste inserts text in the focused field.
func (a *App) Paste(ctx context.Context, text string) error {
	return a.page.Paste(ctx, text)
}

// Cut returns the focused field and clears it.
func (a *App) Cut(ctx context.Context) (string, bool, error) {
	return a.page.Cut(ctx)
}

// Copy returns the focused field without changing it.
func (a *App) Copy(ctx context.Context) (string, bool, error) {
	return a.page.Copy(ctx)
}

// SelectAll selects the whole focused field.
func (a *App) SelectAll(ctx context.Context) error {
	return a.page.SelectAll(ctx)
}

// FormValue returns the stored value of a control id.
func (a *App) FormValue(id string) string {
	return a.page.FormValue(id)
}

// FormChecked reports whether a checkbox or radio is checked.
func (a *App) FormChecked(id string) bool {
	return a.page.FormChecked(id)
}

// FocusedField returns the focused control id, or "".
func (a *App) FocusedField() string {
	return a.page.FocusedField()
}

// FormSelected reports whole-value selection on id.
func (a *App) FormSelected(id string) bool {
	return a.page.FormSelected(id)
}

// Hover sets the hovered element from a point.
func (a *App) Hover(ctx context.Context, x, y float64) error {
	return a.page.Hover(ctx, x, y)
}

// Press sets the pressed element from a point.
func (a *App) Press(ctx context.Context, x, y float64) error {
	return a.page.Press(ctx, x, y)
}

// Release clears the pressed element.
func (a *App) Release(ctx context.Context) error {
	return a.page.Release(ctx)
}
