// Package input is the keyboard focus and pointer interaction example.
// Two fields, a checkbox, a link, and a long list put focus traversal, the
// caret keys, drag selection, the context menu, cursor shapes, touch scroll
// and pinch, page scrolling, and F11 in one window. ownframe opens the window;
// this package does not.
package input

import (
	"context"
	_ "embed"

	"github.com/chinmay-sawant/ownframe"
)

//go:embed input.html
var inputHTML string

const (
	// DefaultWidth and DefaultHeight are the size of a newly opened window.
	DefaultWidth  = 420
	DefaultHeight = 560
	// scrollEnd scrolls past any content length.
	scrollEnd = 1 << 30
)

// App is the input screen.
type App struct {
	page *ownframe.Page
}

// New parses the embedded template.
func New() (*App, error) {
	p, err := ownframe.New(ownframe.Config{DevTools: true,
		Title:  "Input",
		HTML:   inputHTML,
		Width:  DefaultWidth,
		Height: DefaultHeight,
	})
	if err != nil {
		return nil, err
	}

	app := &App{page: p}
	p.Handle(ownframe.Handlers{Click: app.click})

	return app, nil
}

// click scrolls to the end when the link is pressed.
func (a *App) click(_ context.Context, box ownframe.Box) error {
	if box.ID == "more" {
		a.scrollTo(0, scrollEnd)
	}

	return nil
}

func (a *App) scrollTo(x, y int) {
	a.page.ScrollTo(x, y)
}
