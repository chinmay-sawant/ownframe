package eyes

import (
	"math"

	"github.com/chinmay-sawant/ownframe"
	"github.com/chinmay-sawant/ownframe/internal/frame"
)

// fillAt returns the fill laid out at the element box.
func fillAt(d *ownframe.Display, bx []ownframe.Box, id string) *ownframe.DisplayOp {
	var b ownframe.Box
	ok := false
	for _, v := range bx {
		if v.ID == id {
			b = v
			ok = true
			break
		}
	}
	if !ok {
		return nil
	}
	x, y, w, h := frame.BoxUnits(d, b)
	for i := range d.Ops {
		op := &d.Ops[i]
		if op.Kind != ownframe.DisplayOpFillRect {
			continue
		}
		if math.Abs(op.X-x) <= 1.5 && math.Abs(op.Y-y) <= 1.5 &&
			math.Abs(op.W-w) <= 1.5 && math.Abs(op.H-h) <= 1.5 {
			return op
		}
	}
	return nil
}

// textAt returns the first text run inside the box.
func textAt(d *ownframe.Display, bx []ownframe.Box, id string) *ownframe.DisplayOp {
	for _, v := range bx {
		if v.ID == id {
			return frame.Text(d, v)
		}
	}
	return nil
}
