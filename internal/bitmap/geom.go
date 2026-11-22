package bitmap

import "github.com/chinmay-sawant/blinkless/layout"

func transform(op *layout.DisplayOp, x, y float64) (float64, float64) {
	if op.XformSet {
		return op.Transform().Apply(x, y)
	}
	return x, y
}

func untransform(m *layout.Matrix2D, det, x, y float64) (float64, float64) {
	x, y = x-m.E, y-m.F
	return (m.D*x - m.C*y) / det, (-m.B*x + m.A*y) / det
}

func origin(op *layout.DisplayOp) (float64, float64) { return op.X, op.Y }
