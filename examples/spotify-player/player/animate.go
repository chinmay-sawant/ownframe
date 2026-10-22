package player

import (
	"github.com/chinmay-sawant/ownframe"
	"github.com/chinmay-sawant/ownframe/internal/frame"
)

var accent = [3]float64{29.0 / 255, 185.0 / 255, 84.0 / 255}

func (a *App) animate() {
	d := a.page.Display()
	if d == nil {
		return
	}

	boxes := a.page.Boxes()
	pos := a.audio.Position()
	dur := a.audio.Duration()

	var seekBox ownframe.Box
	hasSeek := false
	for _, box := range boxes {
		if box.ID == "seek" {
			seekBox = box
			hasSeek = true
		}
	}

	if dur > 0 {
		if hasSeek {
			fill := accentFill(d, seekBox)
			if fill != nil {
				_, _, w, _ := frame.BoxUnits(d, seekBox)
				fraction := float64(pos) / float64(dur)
				if fraction < 0 {
					fraction = 0
				}

				if fraction > 1 {
					fraction = 1
				}

				oldW := fill.W
				fill.W = w * fraction
				if fill.W != oldW {
					a.page.Invalidate("seek")
				}
			}
		}

		a.animateTimes(d, boxes, pos, dur)
		a.updateView(pos, dur)
	}

	a.animateCredit(d, boxes)
	a.animateEq(d, boxes)
}

func boxByID(boxes []ownframe.Box, id string) (ownframe.Box, bool) {
	for _, box := range boxes {
		if box.ID == id {
			return box, true
		}
	}

	return ownframe.Box{}, false
}

func accentFill(d *ownframe.Display, box ownframe.Box) *ownframe.DisplayOp {
	if d == nil {
		return nil
	}

	x, y, w, h := frame.BoxUnits(d, box)
	var best *ownframe.DisplayOp

	for i := range d.Ops {
		op := &d.Ops[i]
		if op.Kind != ownframe.DisplayOpFillRect {
			continue
		}

		cx, cy := op.X+op.W/centerDiv, op.Y+op.H/centerDiv
		if !(cx >= x-0.5 && cx <= x+w+0.5 && cy >= y-0.5 && cy <= y+h+0.5) {
			continue
		}

		const tolerance = 0.02
		if !(op.Alpha > 0.5 &&
			numAbs(op.R-accent[0]) < tolerance &&
			numAbs(op.G-accent[1]) < tolerance &&
			numAbs(op.B-accent[2]) < tolerance) {
			continue
		}

		if best == nil || op.X < best.X {
			best = op
		}
	}

	return best
}

func numAbs(v float64) float64 {
	if v < 0 {
		return -v
	}

	return v
}
