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

// setLid places one background lid, or parks it offscreen when shut. The
// callers pad the height past the eye edges, so the shut line sits at 3.
func setLid(d *ownframe.Display, op *ownframe.DisplayOp, x, y, w, h float64) {
	if op == nil {
		return
	}
	p := d.PointsPerPixel
	if p == 0 {
		p = 1
	}
	if h < 3 {
		op.X, op.Y, op.W, op.H, op.Alpha = -1000, -1000, 0, 0, 0
		return
	}
	op.X, op.Y, op.W, op.H = x*p, y*p, w*p, h*p
	op.Alpha = 1
}
