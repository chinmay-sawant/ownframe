package main

import (
	"context"
	"strings"

	"github.com/chinmay-sawant/ownframe"
)

// App is the stress dashboard screen. all holds every grid row; view.Rows
// is only the visible window plus overscan.
type App struct {
	page     *ownframe.Page
	view     View
	rows     int
	all      []Row
	winStart int
	winEnd   int
	state    tickState
	track    ownframe.Box
	eq       ownframe.Box
}

// newApp parses the dashboard template and wires input and the tick.
func newApp(rows int) (*App, error) {
	page, err := ownframe.NewWithOptions(ownframe.Config{DevTools: true,
		Title: "Stress", HTML: buildHTML(),
		Width: 1280, Height: 900, MinWidth: 640, MinHeight: 480,
	}, ownframe.WithPerf(true))
	if err != nil {
		return nil, err
	}
	a := &App{page: page, view: makeView(rows), rows: rows}
	a.all = a.view.Rows
	a.winStart, a.winEnd = -1, -1
	page.SetWindowing(true)
	page.SetScrollWindow(a.applyWindow)
	a.applyWindow(0, 900)
	page.Handle(ownframe.Handlers{Click: a.onClick, Change: a.onChange})
	page.SetTick(a.Tick)
	page.SetData(a.view)
	return a, nil
}

// onClick switches the sidebar section or bumps progress, then redraws.
func (a *App) onClick(_ context.Context, box ownframe.Box) error {
	if strings.HasPrefix(box.ID, "nav-") {
		a.view.Active = strings.TrimPrefix(box.ID, "nav-")
	}
	if box.ID == "go" {
		a.view.Progress = (a.view.Progress + 10) % 101
	}
	if box.ID == "reset" {
		a.view = makeView(a.rows)
		a.all = a.view.Rows
		a.winStart, a.winEnd = -1, -1
		_, y := a.page.ScrollOffset()
		a.applyWindow(y, 900)
	}
	a.page.SetData(a.view)
	return nil
}

// onChange keeps the filter value by re-setting the view.
func (a *App) onChange(_ context.Context, _ ownframe.Box) error {
	a.page.SetData(a.view)
	return nil
}
