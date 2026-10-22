package main

import (
	"github.com/chinmay-sawant/ownframe"
)

// bind re-finds the animated ops after a Redraw replaced the display list.
func (a *App) bind(d *ownframe.Display) {
	boxes := a.page.Boxes()
	a.state.seek = nil
	a.state.bars = nil
	a.state.label = nil
	for _, b := range boxes {
		switch b.ID {
		case "seek":
			a.state.seek = ownframe.Fill(d, b, accent)
		case "eq":
			a.eq = b
			a.state.bars = ownframe.Fills(d, b, accent)
		case "ptrack":
			a.track = b
		case "tick":
			a.state.label = ownframe.Text(d, b)
		}
	}
	a.state.bound = a.page.Generation()
}
