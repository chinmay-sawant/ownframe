package shapes

import (
	"context"

	"github.com/chinmay-sawant/ownframe"
)

// Page returns the ownframe page Run and Serve display.
func (a *App) Page() *ownframe.Page {
	return a.page
}

// Replayable reports whether the last draw kept a display list. A false
// result means the page fell back to a bitmap.
func (a *App) Replayable() bool {
	return a.page.Display() != nil
}

// Generation increases by one on every successful Redraw.
func (a *App) Generation() uint64 {
	return a.page.Generation()
}

// SetSize stores the frame size used by the next Redraw.
func (a *App) SetSize(width, height int) {
	a.page.SetSize(width, height)
}

// Redraw fills the template and lays the current size out.
func (a *App) Redraw(ctx context.Context) error {
	return a.page.Redraw(ctx)
}

// PNG returns the last picture as PNG bytes.
func (a *App) PNG() []byte {
	return a.page.PNG()
}

// Boxes returns the last hit-test boxes.
func (a *App) Boxes() []ownframe.Box {
	return a.page.Boxes()
}

// Click hit-tests the page and reports the box under the point.
func (a *App) Click(ctx context.Context, x, y float64) error {
	return a.page.Click(ctx, x, y)
}

// View returns the struct the template prints.
func (a *App) View() *View {
	return &a.view
}
