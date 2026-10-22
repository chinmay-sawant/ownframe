package console

import (
	"context"

	"github.com/chinmay-sawant/ownframe"
)

// Page returns the ownframe page Run and Serve display.
func (a *App) Page() *ownframe.Page {
	return a.page
}

// View returns a copy of the current screen data.
func (a *App) View() View {
	return a.view
}

// Redraw fills the template and renders the current size.
func (a *App) Redraw(ctx context.Context) error {
	return a.page.Redraw(ctx)
}

// Boxes returns the last hit-test boxes.
func (a *App) Boxes() []ownframe.Box {
	return a.page.Boxes()
}

// Click hit-tests the page and applies the data-action.
func (a *App) Click(ctx context.Context, x, y float64) error {
	return a.page.Testing().Click(ctx, x, y)
}

// Type appends text to the focused field.
func (a *App) Type(ctx context.Context, text string) error {
	return a.page.Testing().Type(ctx, text)
}

// Submit runs the login or note form submit.
func (a *App) Submit(ctx context.Context) error {
	return a.page.Testing().Submit(ctx)
}
