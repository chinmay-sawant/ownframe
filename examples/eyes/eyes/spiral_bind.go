package eyes

import (
	"math"

	"github.com/chinmay-sawant/ownframe"
)

func imageAt(d *ownframe.Display, boxes []ownframe.Box, id string) *ownframe.DisplayOp {
	for _, b := range boxes {
		if b.ID != id {
			continue
		}
		x, y, w, h := ownframe.BoxUnits(d, b)
		for i := range d.Ops {
			op := &d.Ops[i]
			data, _, _ := op.ImageBytes()
			if len(data) != 0 && math.Abs(op.X-x) < 1.5 &&
				math.Abs(op.Y-y) < 1.5 && math.Abs(op.W-w) < 1.5 &&
				math.Abs(op.H-h) < 1.5 {
				return op
			}
		}
	}
	return nil
}
