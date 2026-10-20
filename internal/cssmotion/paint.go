package cssmotion

import "github.com/chinmay-sawant/ownframe"

// Paint is the set of display-list values sampled by CSS motion.
type Paint struct {
	X, Y, W, H     float64
	R, G, B, Alpha float64
	Opacity        float64
}

// Read snapshots the paint values from an operation.
func Read(op *ownframe.DisplayOp) Paint {
	return Paint{X: op.X, Y: op.Y, W: op.W, H: op.H,
		R: op.R, G: op.G, B: op.B, Alpha: 1, Opacity: op.Opacity()}
}

// Write applies paint values to an operation.
func Write(op *ownframe.DisplayOp, p Paint) {
	op.X, op.Y, op.W, op.H = p.X, p.Y, p.W, p.H
	op.R, op.G, op.B, op.Alpha = p.R, p.G, p.B, p.Alpha
	op.SetPaintOpacity(p.Opacity)
}

// Blend linearly interpolates every sampled paint value.
func Blend(a, b Paint, t float64) Paint {
	f := func(x, y float64) float64 { return x + (y-x)*t }
	return Paint{X: f(a.X, b.X), Y: f(a.Y, b.Y), W: f(a.W, b.W), H: f(a.H, b.H),
		R: f(a.R, b.R), G: f(a.G, b.G), B: f(a.B, b.B), Alpha: f(a.Alpha, b.Alpha),
		Opacity: f(a.Opacity, b.Opacity)}
}
