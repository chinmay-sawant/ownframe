package frame

import (
	"github.com/chinmay-sawant/blinkless/layout"
	"github.com/chinmay-sawant/ownframe"
)

// write stores m as the op's baked transform.
func (m mat6) write(op *ownframe.DisplayOp) {
	var zero layout.DisplayOp
	x := zero.Transform()
	x.A, x.B, x.C, x.D, x.E, x.F = m[0], m[1], m[2], m[3], m[4], m[5]
	op.SetXform(x)
}

// baseOf is the op's baked transform, or the identity when it carries none.
func baseOf(op *ownframe.DisplayOp) mat6 {
	if !op.XformSet {
		return mat6{1, 0, 0, 1, 0, 0}
	}

	m := op.Transform()

	return mat6{m.A, m.B, m.C, m.D, m.E, m.F}
}

// setScale writes scale s about op's painted center outside its base.
func (a *Anim) setScale(i int, op *ownframe.DisplayOp, s float64) {
	cx, cy := paintCenter(op)
	scaleAbout(s, cx, cy).mul(a.base[i]).write(op)
}

// setRotate writes a rotation by deg about op's painted center outside its
// base.
func (a *Anim) setRotate(i int, op *ownframe.DisplayOp, deg float64) {
	cx, cy := paintCenter(op)
	rotateAbout(deg, cx, cy).mul(a.base[i]).write(op)
}
