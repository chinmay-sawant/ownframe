package ui

// The accessors main and the tests use.

import (
	"context"

	"github.com/chinmay-sawant/ownframe"
)

// Page returns the ownframe page Run displays.
func (a *App) Page() *ownframe.Page { return a.page }

// View returns a copy of the printable view.
func (a *App) View() View { return a.view }

// Redraw fills the template and lays the current view out.
func (a *App) Redraw(ctx context.Context) error {
	a.page.SetData(a.view)

	return a.page.Redraw(ctx)
}

// Boxes returns the hit-test boxes of the last drawing.
func (a *App) Boxes() []ownframe.Box { return a.page.Boxes() }

// Click hit-tests the page and applies the control's action.
func (a *App) Click(ctx context.Context, x, y float64) error {
	return a.page.Click(ctx, x, y)
}

// Close removes the tick and stops the backend workers within its budget.
func (a *App) Close() error {
	a.page.SetTick(nil)

	return a.backend.Close()
}
