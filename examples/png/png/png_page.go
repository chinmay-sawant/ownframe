package png

import (
	"context"
	"errors"
	"os"
	"path/filepath"

	"github.com/chinmay-sawant/ownframe"
)

// Page returns the ownframe page Run and Serve display.
func (a *App) Page() *ownframe.Page {
	return a.page
}

// PNG returns the cached PNG bytes for the last draw. The first call after a
// Redraw rasterizes a display-list page on demand.
func (a *App) PNG() []byte {
	return a.page.PNG()
}

// Redraw fills the template and draws the current size.
func (a *App) Redraw(ctx context.Context) error {
	return a.page.Redraw(ctx)
}

// SavePNG writes the current picture to a.png inside dir and returns the
// full path.
func (a *App) SavePNG(dir string) (string, error) {
	data := a.page.PNG()
	if data == nil {
		return "", errors.New("png: no picture")
	}

	path := filepath.Join(dir, "a.png")
	if err := os.WriteFile(path, data, savedPNGPerm); err != nil {
		return "", err
	}

	return path, nil
}

// Generation increases by one on every successful Redraw.
func (a *App) Generation() uint64 {
	return a.page.Generation()
}

// SetSize stores the frame size used by the next Redraw.
func (a *App) SetSize(width, height int) {
	a.page.SetSize(width, height)
}

// Boxes returns the last hit-test boxes.
func (a *App) Boxes() []ownframe.Box {
	return a.page.Boxes()
}

// Click hit-tests the page and activates the button under the point.
func (a *App) Click(ctx context.Context, x, y float64) error {
	return a.page.Click(ctx, x, y)
}

// View returns the struct the template prints.
func (a *App) View() *View {
	return &a.view
}
