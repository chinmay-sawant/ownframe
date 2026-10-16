// Package watch implements a mechanical skeleton watch with a 28,800 vph
// tourbillon escapement, chronograph, and power reserve indicator.
package watch

import (
	"context"
	"time"

	"github.com/chinmay-sawant/ownframe"
)

// App coordinates the watch movement, display list animation, and inputs.
type App struct {
	page     *ownframe.Page
	view     View
	chrono   Chrono
	reserve  PowerReserve
	lume     bool
	now      func() time.Time
	lastTick time.Time
	bound    uint64
	parts    watchParts
}

// New initializes the skeleton watch window and embeds the dial template.
func New() (*App, error) {
	p, err := ownframe.New(ownframe.Config{
		DevTools:  true,
		Title:     "Manufacture Tourbillon Chronograph",
		HTML:      buildHTML(),
		Width:     840,
		Height:    880,
		MinWidth:  640,
		MinHeight: 680,
	})
	if err != nil {
		return nil, err
	}

	app := &App{
		page:     p,
		view:     defaultView(),
		reserve:  newPowerReserve(48.0),
		now:      time.Now,
		lastTick: time.Now(),
		bound:    ^uint64(0),
	}

	p.Handle(ownframe.Handlers{
		Click:   app.onClick,
		KeyDown: app.onKeyDown,
	})
	p.SetTick(app.Tick)
	p.SetData(&app.view)

	return app, nil
}

// Page returns the underlying ownframe page.
func (a *App) Page() *ownframe.Page {
	return a.page
}

// Redraw synchronizes the template and re-renders the viewport.
func (a *App) Redraw(ctx context.Context) error {
	a.page.SetData(&a.view)
	return a.page.Redraw(ctx)
}
