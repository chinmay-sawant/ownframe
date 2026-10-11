package benchutil

import "github.com/chinmay-sawant/ownframe"

// place moves one fill to x, y, w, h in CSS pixels.
func place(d *ownframe.Display, op *ownframe.DisplayOp, x, y, w, h float64) {
	if op == nil {
		return
	}

	p := d.PointsPerPixel
	op.X, op.Y, op.W, op.H = x*p, y*p, w*p, h*p
	op.Alpha = 1
}

// hide collapses one fill off the canvas. A hidden op keeps its kind, so a
// later place draws it again.
func hide(op *ownframe.DisplayOp) {
	if op == nil {
		return
	}

	op.X, op.Y = -1000, -1000
	op.W, op.H = 0, 0
	op.Alpha = 0
}

// fillIn returns the first fill whose center lies in the element's box.
func fillIn(d *ownframe.Display, boxes []ownframe.Box, id string) *ownframe.DisplayOp {
	for _, b := range boxes {
		if b.ID != id {
			continue
		}

		p := d.PointsPerPixel
		x, y, w, h := b.X*p, b.Y*p, b.W*p, b.H*p

		for i := range d.Ops {
			op := &d.Ops[i]
			if op.Kind != ownframe.DisplayOpFillRect {
				continue
			}

			cx, cy := op.X+op.W/2, op.Y+op.H/2

			if cx >= x-0.5 && cx <= x+w+0.5 && cy >= y-0.5 && cy <= y+h+0.5 {
				return op
			}
		}
	}

	return nil
}

// textIn returns the first text run in the element's box. A text op carries
// its baseline in Y, so the baseline is tested against the box.
func textIn(d *ownframe.Display, boxes []ownframe.Box, id string) *ownframe.DisplayOp {
	for _, b := range boxes {
		if b.ID != id {
			continue
		}

		p := d.PointsPerPixel
		x, y, w, h := b.X*p, b.Y*p, b.W*p, b.H*p

		for i := range d.Ops {
			op := &d.Ops[i]
			if op.Kind != ownframe.DisplayOpText {
				continue
			}

			if op.X >= x-0.5 && op.X <= x+w+0.5 && op.Y >= y-0.5 && op.Y <= y+h+0.5 {
				return op
			}
		}
	}

	return nil
}
