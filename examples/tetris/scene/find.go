package scene

import (
	"math"

	"github.com/chinmay-sawant/ownframe"
)

// fillAt returns the fill laid out exactly at the element's box, so a
// cell is matched by its place instead of its colour.
func fillAt(d *ownframe.Display, boxes []ownframe.Box, id string) *ownframe.DisplayOp {
	box, ok := boxByID(boxes, id)
	if !ok {
		return nil
	}

	x, y, w, h := ownframe.BoxUnits(d, box)

	for i := range d.Ops {
		op := &d.Ops[i]
		if op.Kind != ownframe.DisplayOpFillRect {
			continue
		}

		if near(op.X, x) && near(op.Y, y) && near(op.W, w) && near(op.H, h) {
			return op
		}
	}

	return nil
}

// textAt returns the first text run inside the element's box.
func textAt(d *ownframe.Display, boxes []ownframe.Box, id string) *ownframe.DisplayOp {
	box, ok := boxByID(boxes, id)
	if !ok {
		return nil
	}

	return ownframe.Text(d, box)
}

// boxByID finds one hit-test box by id.
func boxByID(boxes []ownframe.Box, id string) (ownframe.Box, bool) {
	for _, box := range boxes {
		if box.ID == id {
			return box, true
		}
	}

	return ownframe.Box{}, false
}

// near compares two display lengths with a little slack for rounding.
func near(a, b float64) bool {
	const slack = 1.5

	return math.Abs(a-b) <= slack
}
