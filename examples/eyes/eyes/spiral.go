package eyes

import (
	"math"

	"github.com/chinmay-sawant/ownframe"
)

// paintSpiral pops one cached spiral image over each eye.
func (a *App) paintSpiral(d *ownframe.Display, base rect,
	op *ownframe.DisplayOp, mix, angle float64) {
	if d == nil || op == nil || base.w <= 0 || base.h <= 0 {
		return
	}
	if mix < 0.001 {
		setOp(d, op, -100, -100, 0, 0, 0)
		op.XformSet = false
		return
	}
	cx := base.x + base.w/2 + a.st.lookX
	cy := base.y + base.h/2 + a.st.lookY
	sz := min(base.w, base.h) * (0.82 + 0.24*math.Sin(math.Pi*mix))
	setOp(d, op, cx-sz/2, cy-sz/2, sz, sz, mix)
	setOpRotation(op, angle)
}

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
