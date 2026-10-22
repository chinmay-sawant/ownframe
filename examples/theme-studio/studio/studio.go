// Package studio is the theme studio example. The editor on the left holds a
// stylesheet and the preview on the right paints from the custom properties
// that sheet declares. Apply restyles the page and writes the same bytes to
// the theme file, which the page watches, so an edit made in another editor
// lands live. ownframe opens the window. This package does not.
package studio

import (
	"strings"

	"github.com/chinmay-sawant/ownframe"
)

// View is the data the template prints.
type View struct {
	Editor string
	Tokens []Token
	Preset string
	Status string
}

// App is the studio screen.
type App struct {
	page *ownframe.Page
	view View
	path string
}

// studioWidth and friends are the desktop frame the editor opens in.
const (
	studioWidth  = 1100
	studioHeight = 700
	studioMinW   = 820
	studioMinH   = 560
)

// New opens the theme file, seeding it with defaultCSS when it is missing or
// blank, and builds the page with that file as its watched theme.
func New() (*App, error) {
	path, err := themePath()
	if err != nil {
		return nil, err
	}

	css, err := loadTheme(path)
	if err != nil || strings.TrimSpace(css) == "" {
		if err := saveTheme(path, defaultCSS()); err != nil {
			return nil, err
		}

		css, err = loadTheme(path)
		if err != nil {
			return nil, err
		}
	}

	p, err := ownframe.New(ownframe.Config{DevTools: true,
		Title:     "Theme studio",
		HTML:      studioHTML,
		ThemeFile: path,
		Width:     studioWidth,
		Height:    studioHeight,
		MinWidth:  studioMinW,
		MinHeight: studioMinH,
	})
	if err != nil {
		return nil, err
	}

	app := &App{page: p, path: path}
	app.view = View{Editor: css, Tokens: parseTokens(css), Status: "ready"}
	p.Handle(ownframe.Handlers{Click: app.onClick})
	p.SetData(&app.view)

	return app, nil
}
