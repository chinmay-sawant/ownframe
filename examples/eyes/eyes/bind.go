package eyes

import "github.com/chinmay-sawant/ownframe"

// parts caches the display ops the paint step moves.
type parts struct {
	left  *ownframe.DisplayOp
	right *ownframe.DisplayOp
	mood  *ownframe.DisplayOp
	hint  *ownframe.DisplayOp
	dev   *ownframe.DisplayOp
	baseL rect
	baseR rect
	rcL   float64
	rcR   float64
	lidLT *ownframe.DisplayOp
	lidLB *ownframe.DisplayOp
	lidRT *ownframe.DisplayOp
	lidRB *ownframe.DisplayOp
}

// rect is a CSS-pixel box.
type rect struct{ x, y, w, h float64 }

// bind caches eye and text ops after each Redraw or resize.
func (a *App) bind(d *ownframe.Display) {
	bx := a.page.Boxes()
	a.pm.left = fillAt(d, bx, "eye-left")
	a.pm.right = fillAt(d, bx, "eye-right")
	a.pm.mood = textAt(d, bx, "t-mood")
	a.pm.hint = textAt(d, bx, "t-hint")
	a.pm.dev = textAt(d, bx, "dev-toggle")
	a.pm.baseL = baseOf(d, a.pm.left)
	a.pm.baseR = baseOf(d, a.pm.right)
	a.pm.rcL = baseRCss(d, a.pm.left)
	a.pm.rcR = baseRCss(d, a.pm.right)
	a.pm.lidLT = fillAt(d, bx, "lid-lt")
	a.pm.lidLB = fillAt(d, bx, "lid-lb")
	a.pm.lidRT = fillAt(d, bx, "lid-rt")
	a.pm.lidRB = fillAt(d, bx, "lid-rb")
	a.bnd = a.page.Generation()
}

func baseRad(op *ownframe.DisplayOp) (float64, float64) {
	if op == nil {
		return 0, 0
	}
	return op.Radius, op.RadiusY
}

// baseRCss reads the uniform corner radius in CSS pixels: the largest of
// the uniform and per-corner values, so a shorthand and a longhand agree.
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
