package eyes

import "github.com/chinmay-sawant/ownframe"

// setOp places one fill.
func setOp(d *ownframe.Display, op *ownframe.DisplayOp, x, y, w, h, alpha float64) {
	if op == nil {
		return
	}
	p := d.PointsPerPixel
	if p == 0 {
		p = 1
	}
	op.X, op.Y, op.W, op.H = x*p, y*p, w*p, h*p
	op.Alpha = alpha
}
