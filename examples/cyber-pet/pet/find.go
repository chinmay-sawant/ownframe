package pet

import (
	"math"

	"github.com/chinmay-sawant/ownframe"
	"github.com/chinmay-sawant/ownframe/internal/frame"
)

type rect struct {
	x, y, w, h float64
}

func fillAt(d *ownframe.Display, bx []ownframe.Box, id string) *ownframe.DisplayOp {
	if d == nil {
		return nil
	}
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
		if math.Abs(op.X-x) <= 2.0 && math.Abs(op.Y-y) <= 2.0 &&
			math.Abs(op.W-w) <= 2.0 && math.Abs(op.H-h) <= 2.0 {
			return op
		}
	}
	return nil
}

func textAt(d *ownframe.Display, bx []ownframe.Box, id string) *ownframe.DisplayOp {
	if d == nil {
		return nil
	}
	for _, v := range bx {
		if v.ID != id {
			continue
		}
		x, y, w, h := frame.BoxUnits(d, v)
		for i := range d.Ops {
			op := &d.Ops[i]
			if op.Kind == ownframe.DisplayOpText &&
				op.X >= x-2.0 && op.X <= x+w+2.0 &&
				op.Y >= y-2.0 && op.Y <= y+h+2.0 {
				return op
			}
		}
	}
	return nil
}
