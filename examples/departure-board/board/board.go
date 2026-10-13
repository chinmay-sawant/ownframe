// Package board is the split-flap departure board example. Ownframe opens the
// window; this package builds the page and animates its display list from the
// frame tick. A livery stylesheet paints the board, so a click can switch it
// without another parse.
package board

import (
	"time"

	"github.com/chinmay-sawant/ownframe"
)

// App is the departure board screen.
type App struct {
	page       *ownframe.Page
	view       View
	flip       flipper
	livery     Livery
	night      bool
	paused     bool
	feed       int
	lastRotate time.Time
}

// New parses the embedded template, starts on the amber livery, and installs
// the click handler and the frame tick.
func New() (*App, error) {
	p, err := ownframe.New(ownframe.Config{DevTools: true,
		Title:     "Departure board",
		HTML:      boardHTML,
		Width:     1000,
		Height:    640,
		MinWidth:  720,
		MinHeight: 480,
	})
	if err != nil {
		return nil, err
	}

	a := &App{page: p, view: newView(0), lastRotate: time.Now()}
	p.Handle(ownframe.Handlers{Click: a.onClick})
	p.SetTick(a.Tick)

	if err := a.SetLivery(DefaultLivery); err != nil {
		return nil, err
	}

	p.SetData(&a.view)

	return a, nil
}
