package snip

import (
	"context"

	"github.com/chinmay-sawant/ownframe"
)

// App is one snipping-tool screen. shot is the full capture, crop is
// the preview the template shows and save/copy use, and sel is the
// crop rect. ownframe opens the window. This package does not.
type App struct {
	page     *ownframe.Page
	view     View
	shot     []byte
	crop     []byte
	sel      Rect
	anns     []Annotation
	capturer Capturer
	saveDir  string
	web      bool
	demo     bool
}

// New parses the template and takes the first shot with capturer, or
// the OS tools when none is given. A failed capture falls back to
// DemoImage so headless runs keep working.
func New(capturer ...Capturer) (*App, error) {
	page, err := ownframe.New(ownframe.Config{
		Title:  "Snipping tool",
		HTML:   templateHTML,
		Width:  DefaultWidth,
		Height: DefaultHeight,
	})
	if err != nil {
		return nil, err
	}
	app := &App{page: page}
	if len(capturer) > 0 {
		app.capturer = capturer[0]
	}
	page.Handle(ownframe.Handlers{Click: app.onClick, Change: app.onChange})
	page.SetData(&app.view)
	app.capture(context.Background())
	return app, nil
}

// Page returns the page Run and Serve display.
func (a *App) Page() *ownframe.Page { return a.page }

// SetSize stores the frame size used by the next Redraw.
func (a *App) SetSize(w, h int) { a.page.SetSize(w, h) }

// Redraw draws the current view.
func (a *App) Redraw(ctx context.Context) error { return a.page.Redraw(ctx) }

// Preview returns the cropped bytes the template shows.
func (a *App) Preview() []byte { return a.crop }
