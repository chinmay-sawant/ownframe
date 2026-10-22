package platform

import (
	"context"

	"github.com/chinmay-sawant/ownframe"
)

// New parses the embedded platform template and registers its click handler.
func New() (*App, error) {
	page, err := ownframe.New(ownframe.Config{DevTools: true,
		Title:  "Platform",
		HTML:   platformHTML,
		Width:  DefaultWidth,
		Height: DefaultHeight,
	})
	if err != nil {
		return nil, err
	}

	app := &App{page: page}
	page.Handle(ownframe.Handlers{Click: app.onClick})
	page.SetData(app.view)

	return app, nil
}

// onClick bumps the counter the template prints.
func (a *App) onClick(_ context.Context, box ownframe.Box) error {
	if box.ID != "inc" {
		return nil
	}

	a.view.Count++
	a.page.SetData(a.view)
	a.page.Invalidate("count")

	return nil
}
