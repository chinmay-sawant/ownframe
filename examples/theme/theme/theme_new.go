package theme

import (
	"context"

	"github.com/chinmay-sawant/ownframe"
)

// New parses the embedded template, applies the light theme, and installs
// the toggle click handler.
func New() (*App, error) {
	p, err := ownframe.New(ownframe.Config{DevTools: true,
		Title:  "Live theme",
		HTML:   themeHTML,
		Theme:  lightTheme,
		Width:  DefaultWidth,
		Height: DefaultHeight,
	})
	if err != nil {
		return nil, err
	}

	app := &App{page: p}
	app.view = View{Title: "Live theme", Status: "light theme"}
	p.Handle(ownframe.Handlers{Click: app.onClick})
	p.SetData(&app.view)

	return app, nil
}

// onClick switches the extra stylesheet. Click draws the page after the
// handler returns, so the new theme appears without an extra Redraw.
func (a *App) onClick(_ context.Context, box ownframe.Box) error {
	if box.ID != "toggle" {
		return nil
	}

	a.dark = !a.dark

	css, name := lightTheme, "light theme"
	if a.dark {
		css, name = darkTheme, "dark theme"
	}

	if err := a.page.SetTheme(css); err != nil {
		return err
	}

	a.view.Status = name

	return nil
}
