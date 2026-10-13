// Package app is the Wispr Flow app shell. One HTML template holds the
// collapsible sidebar and the page fragment each screen package provides.
// View carries every page's data, and a nav click switches pages. ownframe
// opens the window. This package does not.
package app

import "github.com/chinmay-sawant/ownframe"

const (
	// DefaultWidth and DefaultHeight are the size of a newly opened window.
	DefaultWidth  = 1638
	DefaultHeight = 950

	// MinWidth and MinHeight are the smallest frame the screen will draw.
	MinWidth  = 480
	MinHeight = 560
)

// App is the Wispr Flow app screen.
type App struct {
	page   *ownframe.Page
	view   View
	streak int
}

// New parses the app template, registers its images and handlers.
func New() (*App, error) {
	page, err := ownframe.New(ownframe.Config{DevTools: true,
		Title:     "Flow",
		HTML:      buildHTML(),
		Width:     DefaultWidth,
		Height:    DefaultHeight,
		MinWidth:  MinWidth,
		MinHeight: MinHeight,
	})
	if err != nil {
		return nil, err
	}

	app := &App{page: page, view: DefaultView()}
	registerImages(page)
	page.Handle(ownframe.Handlers{Click: app.onClick})
	page.SetData(app.view)

	return app, nil
}
