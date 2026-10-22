package history

import (
	"context"
	"errors"

	"github.com/chinmay-sawant/ownframe"
)

// New parses the embedded history template, registers its routes, and
// installs the click handler.
func New() (*App, error) {
	p, err := ownframe.New(ownframe.Config{DevTools: true,
		Title:  "History",
		HTML:   historyHTML,
		Width:  DefaultWidth,
		Height: DefaultHeight,
	})
	if err != nil {
		return nil, err
	}

	app := &App{page: p}
	app.view.Status = "html has initial"
	if err := p.Route("red", routePage("RED", "#b91c1c")); err != nil {
		return nil, err
	}
	if err := p.Route("green", routePage("GREEN", "#176b45")); err != nil {
		return nil, err
	}
	p.Handle(ownframe.Handlers{Click: app.onClick})
	p.SetData(&app.view)

	return app, nil
}

// onClick handles the toolbar boxes that are not data-action routes.
// A routed box loads its page inside Click and never reaches this function.
func (a *App) onClick(ctx context.Context, box ownframe.Box) error {
	switch box.ID {
	case "back":
		return a.step(ctx, a.page.Back)
	case "forward":
		return a.step(ctx, a.page.Forward)
	case "native":
		if err := a.page.Load(ctx, nativeHTML); err != nil {
			return err
		}

		a.view.Status = "loaded " + markerIn(a.page.HTML())

		return nil
	}

	return nil
}

// step runs a history move and records a missing entry in the status line.
func (a *App) step(ctx context.Context, move func(context.Context) error) error {
	err := move(ctx)
	if errors.Is(err, ownframe.ErrNoHistory) {
		a.view.Status = "no history"

		return nil
	}

	if err != nil {
		return err
	}

	a.view.Status = "html has " + markerIn(a.page.HTML())

	return nil
}
