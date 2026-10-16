package eyes

import "github.com/chinmay-sawant/ownframe"

func baseOf(d *ownframe.Display, op *ownframe.DisplayOp) rect {
	if op == nil || d == nil {
		return rect{}
	}
	p := d.PointsPerPixel
	if p == 0 {
		p = 1
	}
	return rect{x: op.X / p, y: op.Y / p, w: op.W / p, h: op.H / p}
}
