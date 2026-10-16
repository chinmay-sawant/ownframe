package watch

import (
	"math"

	"github.com/chinmay-sawant/ownframe"
)

// rotateHand rotates op about its bottom-center pivot (50% 100%).
func rotateHand(op *ownframe.DisplayOp, deg float64) {
	if op == nil {
		return
	}
	px := op.X + op.W/2.0
	py := op.Y + op.H
	applyRotationPivot(op, deg, px, py)
}

// rotateCenter rotates op about its geometric center (50% 50%).
func rotateCenter(op *ownframe.DisplayOp, deg float64) {
	if op == nil {
		return
	}
	cx := op.X + op.W/2.0
	cy := op.Y + op.H/2.0
	applyRotationPivot(op, deg, cx, cy)
}

// applyRotationPivot sets the affine transform matrix rotating around (px, py).
func applyRotationPivot(op *ownframe.DisplayOp, deg, px, py float64) {
	r := deg * math.Pi / 180.0
	c, s := math.Cos(r), math.Sin(r)
	var zero ownframe.DisplayOp
	m := zero.Transform()
	m.A, m.B, m.C, m.D = c, s, -s, c
	m.E = px - c*px + s*py
	m.F = py - s*px - c*py
	op.SetXform(m)
}
