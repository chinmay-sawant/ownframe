package frame

import (
	"math"
)

// mat6 is a 2D affine transform: x' = a*x + c*y + e, y' = b*x + d*y + f.
//
// The public layout package does not export the engine's matrix type, so
// this package composes its own and writes the fields through the value
// op.Transform returns. The engine's type cannot be named outside it, so
// the write in anim_write.go is the only place the two meet.
type mat6 [6]float64

// scaleAbout is a scale by s around (cx, cy).
func scaleAbout(s, cx, cy float64) mat6 {
	return mat6{s, 0, 0, s, cx * (1 - s), cy * (1 - s)}
}

// rotateAbout is a rotation by deg degrees around (cx, cy).
func rotateAbout(deg, cx, cy float64) mat6 {
	r := deg * math.Pi / 180
	c, s := math.Cos(r), math.Sin(r)

	return mat6{c, s, -s, c, cx - c*cx + s*cy, cy - s*cx - c*cy}
}

// mul returns m * n, applying n first, like the engine's Mul.
func (m mat6) mul(n mat6) mat6 {
	return mat6{
		m[0]*n[0] + m[2]*n[1],
		m[1]*n[0] + m[3]*n[1],
		m[0]*n[2] + m[2]*n[3],
		m[1]*n[2] + m[3]*n[3],
		m[0]*n[4] + m[2]*n[5] + m[4],
		m[1]*n[4] + m[3]*n[5] + m[5],
	}
}
