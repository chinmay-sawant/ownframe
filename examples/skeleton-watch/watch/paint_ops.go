package watch

import (
	"math"

	"github.com/chinmay-sawant/ownframe"
)

type mat6 [6]float64

// partAssembly groups all display ops of an animated watch component.
type partAssembly struct {
	ops  []*ownframe.DisplayOp
	base []mat6
	px   float64
	py   float64
}

// rotate sets the combined affine transform matrix rotating around (px, py).
func (p *partAssembly) rotate(deg float64) {
	if len(p.ops) == 0 {
		return
	}
	r := deg * math.Pi / 180.0
	c, s := math.Cos(r), math.Sin(r)
	rot := mat6{c, s, -s, c, p.px - c*p.px + s*p.py, p.py - s*p.px - c*p.py}
	var zero ownframe.DisplayOp
	for i, op := range p.ops {
		m := mulMat(rot, p.base[i])
		x := zero.Transform()
		x.A, x.B, x.C, x.D, x.E, x.F = m[0], m[1], m[2], m[3], m[4], m[5]
		op.SetXform(x)
	}
}

func baseOf(op *ownframe.DisplayOp) mat6 {
	if !op.XformSet {
		return mat6{1, 0, 0, 1, 0, 0}
	}
	m := op.Transform()
	return mat6{m.A, m.B, m.C, m.D, m.E, m.F}
}

func mulMat(m, n mat6) mat6 {
	return mat6{
		m[0]*n[0] + m[2]*n[1],
		m[1]*n[0] + m[3]*n[1],
		m[0]*n[2] + m[2]*n[3],
		m[1]*n[2] + m[3]*n[3],
		m[0]*n[4] + m[2]*n[5] + m[4],
		m[1]*n[4] + m[3]*n[5] + m[5],
	}
}
