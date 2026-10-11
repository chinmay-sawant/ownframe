package bitmap

import "github.com/chinmay-sawant/blinkless/layout"

func transform(op *layout.DisplayOp, x, y float64) (float64, float64) {
	if op.XformSet {
		return op.Transform().Apply(x, y)
	}
	return x, y
}

func untransform(op *layout.DisplayOp, x, y float64) (float64, float64, bool) {
	if !op.XformSet {
		return x, y, true
	}
	m := op.Transform()
	det := m.A*m.D - m.B*m.C
	if det == 0 {
		return 0, 0, false
	}
	x, y = x-m.E, y-m.F
	return (m.D*x - m.C*y) / det, (-m.B*x + m.A*y) / det, true
}

func origin(op *layout.DisplayOp) (float64, float64) { return op.X, op.Y }
