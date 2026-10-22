package devtools

import (
	"context"

	"github.com/chinmay-sawant/ownframe"
)

// Page returns the ownframe page Run and Serve display.
func (a *App) Page() *ownframe.Page {
	return a.page
}

// View returns the counter.
func (a *App) View() View {
	return a.view
}

// Redraw fills the template and renders the current size.
func (a *App) Redraw(ctx context.Context) error {
	return a.page.Redraw(ctx)
}

// Display returns the last display list, or nil on a fallback page.
func (a *App) Display() *ownframe.Display {
	return a.page.Display()
}

// Boxes returns the last hit-test boxes.
func (a *App) Boxes() []ownframe.Box {
	return a.page.Boxes()
}

// Click hit-tests the page and applies the box under the point.
func (a *App) Click(ctx context.Context, x, y float64) error {
	return a.page.Click(ctx, x, y)
}
