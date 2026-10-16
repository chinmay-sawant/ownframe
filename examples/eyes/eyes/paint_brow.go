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

func paintBrow(d *ownframe.Display, op *ownframe.DisplayOp, b rect,
	style eyeStyle, lookX, lookY float64) {
	if op == nil || style.browUp <= 0 {
		if op != nil {
			op.Alpha = 0
		}
		return
	}
	w, h := b.w*.68, 8.0
	x := b.x + lookX + style.gazeX + (b.w-w)/2
	y := b.y + lookY + style.gazeY - style.browUp - h - 8
	setOp(d, op, x, y, w, h, 1)
	p := d.PointsPerPixel
	if p == 0 {
		p = 1
	}
	op.Radius = h / 2 * p
	setOpRotation(op, style.browSlant)
}
