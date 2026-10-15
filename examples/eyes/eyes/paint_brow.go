package eyes

import (
	"math"

	"github.com/chinmay-sawant/ownframe"
)

func setOpRotation(op *ownframe.DisplayOp, deg float64) {
	if op == nil {
		return
	}
	if deg == 0 {
		op.XformSet = false
		return
	}
	r := deg * math.Pi / 180
	c, s := math.Cos(r), math.Sin(r)
	cx, cy := op.X+op.W/2, op.Y+op.H/2
	var zero ownframe.DisplayOp
	m := zero.Transform()
	m.A, m.B, m.C, m.D = c, s, -s, c
	m.E, m.F = cx-c*cx+s*cy, cy-s*cx-c*cy
	op.SetXform(m)
}
