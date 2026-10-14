package bind

import (
	"context"

	"github.com/chinmay-sawant/ownframe"
)

// New parses the embedded bind template and registers its change handler.
func New() (*App, error) {
	page, err := ownframe.New(ownframe.Config{DevTools: true,
		Title:  "Bind",
		HTML:   bindHTML,
		Width:  DefaultWidth,
		Height: DefaultHeight,
	})
	if err != nil {
		return nil, err
	}

	app := &App{page: page}
	app.view.Color = "Red"
	page.Handle(ownframe.Handlers{Change: app.onChange})
	page.SetData(&app.view)

	return app, nil
}

// onChange records which bound control changed. The library calls it after a
// control changes and before the page redraws, so the status line follows.
func (a *App) onChange(_ context.Context, box ownframe.Box) error {
	a.view.Status = "changed " + box.ID

	return nil
}
