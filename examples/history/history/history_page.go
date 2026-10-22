package history

import (
	"context"

	"github.com/chinmay-sawant/ownframe"
)

// Page returns the ownframe page Run and Serve display.
func (a *App) Page() *ownframe.Page { return a.page }

// View returns the current status line data.
func (a *App) View() *View { return &a.view }

// Redraw fills the template and renders the current size.
func (a *App) Redraw(ctx context.Context) error { return a.page.Redraw(ctx) }

// PNG returns the last PNG, or nil when nothing has been drawn.
func (a *App) PNG() []byte { return a.page.PNG() }

// Boxes returns the last hit-test boxes.
func (a *App) Boxes() []ownframe.Box { return a.page.Boxes() }

// Click hit-tests the page and applies the click. A data-action box with a
// registered route loads that page and skips the click handler.
func (a *App) Click(ctx context.Context, x, y float64) error {
	return a.page.Click(ctx, x, y)
}

// Back draws the previous template source.
func (a *App) Back(ctx context.Context) error { return a.page.Back(ctx) }

// Forward draws the next template source.
func (a *App) Forward(ctx context.Context) error { return a.page.Forward(ctx) }

// Load parses html, drops later history, and draws that page.
func (a *App) Load(ctx context.Context, html string) error {
	return a.page.Load(ctx, html)
}

// HTML returns the source at the current history index.
func (a *App) HTML() string { return a.page.HTML() }
