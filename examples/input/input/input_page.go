package input

import (
	"context"

	"github.com/chinmay-sawant/ownframe"
)

// Page returns the ownframe page Run and Serve display.
func (a *App) Page() *ownframe.Page { return a.page }

// SetSize stores the frame size used by the next Redraw.
func (a *App) SetSize(width, height int) { a.page.SetSize(width, height) }

// Size returns the frame size in CSS pixels.
func (a *App) Size() (int, int) { return a.page.Size() }

// Redraw fills the template and renders the current size.
func (a *App) Redraw(ctx context.Context) error { return a.page.Redraw(ctx) }

// Boxes returns the last hit-test boxes.
func (a *App) Boxes() []ownframe.Box { return a.page.Boxes() }

// Click hit-tests the page and applies the click.
func (a *App) Click(ctx context.Context, x, y float64) error {
	return a.page.Click(ctx, x, y)
}

// Type types into the focused field.
func (a *App) Type(ctx context.Context, text string) error {
	return a.page.Type(ctx, text)
}

// FocusedField returns the id of the focused control, or "".
func (a *App) FocusedField() string { return a.page.FocusedField() }

// FormValue returns the value of a control.
func (a *App) FormValue(id string) string { return a.page.FormValue(id) }

// FormChecked returns the checked state of a checkbox.
func (a *App) FormChecked(id string) bool { return a.page.FormChecked(id) }
