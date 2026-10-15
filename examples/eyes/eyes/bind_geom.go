package eyes

import "github.com/chinmay-sawant/ownframe"

func baseRCss(d *ownframe.Display, op *ownframe.DisplayOp) float64 {
	if op == nil || d == nil {
		return 0
	}
	p := d.PointsPerPixel
	if p == 0 {
		p = 1
	}
	r := op.Radius
	for _, c := range []float64{op.RadiusTopLeft, op.RadiusTopRight,
		op.RadiusBottomRight, op.RadiusBottomLeft} {
		r = max(r, c)
	}
	return r / p
}

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
