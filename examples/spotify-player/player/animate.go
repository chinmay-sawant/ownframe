package player

import "github.com/chinmay-sawant/ownframe"

// accent is Spotify green, #1db954.
var accent = [3]float64{29.0 / 255, 185.0 / 255, 84.0 / 255}

// animate moves the seek bar, the clock, the credit, and the equalizer
// without another Redraw.
func (a *App) animate() {
	d := a.page.Display()
	if d == nil {
		return
	}

	boxes := a.page.Boxes()
	pos := a.audio.Position()
	dur := a.audio.Duration()

	if dur > 0 {
		a.animateSeek(d, boxes, pos, dur)
		a.animateTimes(d, boxes, pos, dur)
		a.updateView(pos, dur)
	}

	a.animateCredit(d, boxes)
	a.animateEq(d, boxes)
}

// boxByID returns the first hit-test box with id.
func boxByID(boxes []ownframe.Box, id string) (ownframe.Box, bool) {
	for _, box := range boxes {
		if box.ID == id {
			return box, true
		}
	}

	return ownframe.Box{}, false
}
