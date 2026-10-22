package resize

import (
	"context"

	"github.com/chinmay-sawant/ownframe"
)

// Page returns the ownframe page Run and Serve display.
func (a *App) Page() *ownframe.Page {
	return a.page
}

// SetSize stores the frame size used by the next Redraw.
func (a *App) SetSize(width, height int) {
	a.page.SetSize(width, height)
}

// Redraw fills the template and lays the current size out.
func (a *App) Redraw(ctx context.Context) error {
	return a.page.Redraw(ctx)
}

// Boxes returns the last hit-test boxes.
func (a *App) Boxes() []ownframe.Box {
	return a.page.Boxes()
}

// Hover sets the hovered element from a point.
func (a *App) Hover(ctx context.Context, x, y float64) error {
	return a.page.Hover(ctx, x, y)
}

// Generation increases by one on every successful Redraw.
func (a *App) Generation() uint64 {
	return a.page.Generation()
}
