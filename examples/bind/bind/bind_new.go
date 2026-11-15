package bind

import (
	"context"
	"fmt"

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
	page.Handle(ownframe.Handlers{
		Click:  app.onClick,
		Change: app.onChange,
	})
	page.SetData(&app.view)

	return app, nil
}

func (a *App) onClick(_ context.Context, box ownframe.Box) error {
	if box.Action != "send" {
		return nil
	}

	a.view.Status = fmt.Sprintf(
		"sent name=%s email=%s agree=%t plan=%s color=%s",
		a.view.Name,
		a.view.Email,
		a.view.Agree,
		a.view.Plan,
		a.view.Color,
	)
	a.page.SetData(&a.view)

	return nil
}

// onChange records which bound control changed. The library calls it after a
// control changes and before the page redraws, so the status line follows.
func (a *App) onChange(_ context.Context, box ownframe.Box) error {
	a.view.Status = "changed " + box.ID

	return nil
}
