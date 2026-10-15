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
