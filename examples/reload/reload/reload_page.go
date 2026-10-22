package reload

import (
	"context"

	"github.com/chinmay-sawant/ownframe"
)

// Page returns the ownframe page Run and Serve display.
func (a *App) Page() *ownframe.Page { return a.page }

// View returns the current counter data.
func (a *App) View() *View { return &a.view }

// HTML returns the page source the window shows.
func (a *App) HTML() string { return a.page.HTML() }

// PollReload rereads the source file and reports whether the page changed.
func (a *App) PollReload(ctx context.Context) (bool, error) {
	return a.page.PollReload(ctx)
}

// Redraw fills the template and renders the current size.
func (a *App) Redraw(ctx context.Context) error { return a.page.Redraw(ctx) }

// Boxes returns the last hit-test boxes.
func (a *App) Boxes() []ownframe.Box { return a.page.Boxes() }

// Click hit-tests the page and applies the click.
func (a *App) Click(ctx context.Context, x, y float64) error {
	return a.page.Click(ctx, x, y)
}
