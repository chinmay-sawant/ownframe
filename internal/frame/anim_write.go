package frame

import (
	"github.com/chinmay-sawant/blinkless/layout"
	"github.com/chinmay-sawant/ownframe/internal/page"
)

// write stores m as the op's baked transform. A nil op does nothing.
func (m mat6) write(op *page.DisplayOp) {
	if op == nil {
		return
	}
	var zero layout.DisplayOp
	x := zero.Transform()
	x.A, x.B, x.C, x.D, x.E, x.F = m[0], m[1], m[2], m[3], m[4], m[5]
	op.SetXform(x)
}

// baseOf is the op's baked transform, or the identity when it carries
// none or is nil.
func baseOf(op *page.DisplayOp) mat6 {
	if op == nil {
		return mat6{1, 0, 0, 1, 0, 0}
	}
	if !op.XformSet {
		return mat6{1, 0, 0, 1, 0, 0}
	}

	m := op.Transform()

	return mat6{m.A, m.B, m.C, m.D, m.E, m.F}
}

// setScale writes scale s about op's painted center outside its base.
// A nil op does nothing.
func (a *Anim) setScale(i int, op *page.DisplayOp, s float64) {
	if op == nil {
		return
	}
	cx, cy := paintCenter(op)
	scaleAbout(s, cx, cy).mul(a.base[i]).write(op)
}

// setRotate writes a rotation by deg about op's painted center outside its
// base. A nil op does nothing.
func (a *Anim) setRotate(i int, op *page.DisplayOp, deg float64) {
	if op == nil {
		return
	}
	cx, cy := paintCenter(op)
	rotateAbout(deg, cx, cy).mul(a.base[i]).write(op)
}
