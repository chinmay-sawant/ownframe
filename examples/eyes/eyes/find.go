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
	ops := textOpsAt(d, bx, id)
	if len(ops) > 0 {
		return ops[0]
	}
	return nil
}

func textOpsAt(d *ownframe.Display, bx []ownframe.Box, id string) []*ownframe.DisplayOp {
	if d == nil {
		return nil
	}
	for _, v := range bx {
		if v.ID != id {
			continue
		}
		x, y, w, h := frame.BoxUnits(d, v)
		var ops []*ownframe.DisplayOp
		for i := range d.Ops {
			op := &d.Ops[i]
			if op.Kind == ownframe.DisplayOpText && op.X >= x-0.5 &&
				op.X <= x+w+0.5 && op.Y >= y-0.5 && op.Y <= y+h+0.5 {
				ops = append(ops, op)
			}
		}
		return ops
	}
	return nil
}
