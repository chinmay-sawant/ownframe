package states

import "context"

// Click hit-tests the page and applies the click.
func (a *App) Click(ctx context.Context, x, y float64) error {
	return a.page.Click(ctx, x, y)
}

// Hover sets the hovered element from a point and redraws when it changed.
func (a *App) Hover(ctx context.Context, x, y float64) error {
	return a.page.Hover(ctx, x, y)
}

// Press sets the pressed element from a point and redraws when it changed.
func (a *App) Press(ctx context.Context, x, y float64) error {
	return a.page.Press(ctx, x, y)
}

// Release clears the pressed element and redraws when it changed.
func (a *App) Release(ctx context.Context) error { return a.page.Release(ctx) }

// Type edits the focused control.
func (a *App) Type(ctx context.Context, text string) error {
	return a.page.Type(ctx, text)
}

// FocusedField returns the focused control id, or "" when none is focused.
func (a *App) FocusedField() string { return a.page.FocusedField() }

// FormChecked reports whether the checkbox id is checked.
func (a *App) FormChecked(id string) bool { return a.page.FormChecked(id) }
