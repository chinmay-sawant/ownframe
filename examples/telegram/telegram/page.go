package telegram

import (
	"context"

	"github.com/chinmay-sawant/ownframe"
)

// Page returns the ownframe page Run and BindMobile display.
func (a *App) Page() *ownframe.Page {
	return a.page
}

// View returns a copy of the current view.
func (a *App) View() View {
	return a.view
}

// Redraw fills the template and renders the current size.
func (a *App) Redraw(ctx context.Context) error {
	return a.page.Redraw(ctx)
}

// PNG returns the last PNG, or nil when nothing has been drawn.
func (a *App) PNG() []byte {
	return a.page.PNG()
}

// Boxes returns the last hit-test boxes.
func (a *App) Boxes() []ownframe.Box {
	return a.page.Boxes()
}

// Click hit-tests the page and applies the clicked action.
func (a *App) Click(ctx context.Context, x, y float64) error {
	return a.page.Click(ctx, x, y)
}

// Focus focuses one control, so a test can type without a click.
func (a *App) Focus(ctx context.Context, id string) error {
	return a.page.Focus(ctx, id)
}

// Type sends text to the focused control.
func (a *App) Type(ctx context.Context, text string) error {
	return a.page.Type(ctx, text)
}

// Submit sends the focused composer.
func (a *App) Submit(ctx context.Context) error {
	return a.page.Submit(ctx)
}
