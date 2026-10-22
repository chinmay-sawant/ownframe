package drop

import (
	"context"

	"github.com/chinmay-sawant/ownframe"
)

// Page returns the ownframe page Run and Serve display.
func (a *App) Page() *ownframe.Page {
	return a.page
}

// Drop offers files to the page, the way the window does.
func (a *App) Drop(ctx context.Context, files []ownframe.Drop) error {
	return a.page.Drop(ctx, files)
}

// Click sends one click at the point, the way the window does.
func (a *App) Click(ctx context.Context, x, y float64) error {
	return a.page.Click(ctx, x, y)
}

// View returns the struct the template prints.
func (a *App) View() *View {
	return &a.view
}

// SetSize stores the frame size used by the next Redraw.
func (a *App) SetSize(width, height int) {
	a.page.SetSize(width, height)
}

// Redraw fills the template and draws the current size.
func (a *App) Redraw(ctx context.Context) error {
	return a.page.Redraw(ctx)
}

// PNG returns the cached PNG bytes for the last draw.
func (a *App) PNG() []byte {
	return a.page.PNG()
}

// Boxes returns the last hit-test boxes.
func (a *App) Boxes() []ownframe.Box {
	return a.page.Boxes()
}
