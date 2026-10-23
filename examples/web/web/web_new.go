package web

import (
	"context"

	"github.com/chinmay-sawant/ownframe"
)

// New parses the embedded template and registers its click and change handlers.
func New() (*App, error) {
	page, err := ownframe.New(ownframe.Config{DevTools: true,
		Title:  "Web mode",
		HTML:   webHTML,
		Width:  DefaultWidth,
		Height: DefaultHeight,
	})
	if err != nil {
		return nil, err
	}

	app := &App{page: page}
	page.Handle(ownframe.Handlers{Click: app.onClick, Change: app.onChange})
	page.SetData(&app.view)

	return app, nil
}

// onClick counts the increments and resets them.
func (a *App) onClick(_ context.Context, box ownframe.Box) error {
	switch box.ID {
	case "inc":
		a.view.Count++
	case "reset":
		a.view.Count = 0
	}

	return nil
}

// onChange mirrors the note field into View so the template prints it.
func (a *App) onChange(_ context.Context, box ownframe.Box) error {
	if box.ID == "note" {
		a.view.Note = a.page.Form().FormValue("note")
	}

	return nil
}
