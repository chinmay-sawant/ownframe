package scrolling

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

// Generation increases by one on every successful Redraw.
func (a *App) Generation() uint64 { return a.page.Generation() }

// Redraw fills the template and renders the current size.
func (a *App) Redraw(ctx context.Context) error { return a.page.Redraw(ctx) }

// PNG returns the last PNG, or nil when nothing has been drawn.
func (a *App) PNG() []byte { return a.page.PNG() }

// Boxes returns the last hit-test boxes.
func (a *App) Boxes() []ownframe.Box { return a.page.Boxes() }

// Click hit-tests the page and applies the click.
func (a *App) Click(ctx context.Context, x, y float64) error {
	return a.page.Click(ctx, x, y)
}
