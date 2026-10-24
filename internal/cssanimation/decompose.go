package cssanimation

import (
	"math"

	"github.com/chinmay-sawant/ownframe/internal/animation"
)

// Indexes into the matrix argument lists: matrix() translation and the
// matrix3d() translation column, plus the perspective divisor slot.
const (
	matrixE      = 4
	matrixF      = 5
	matrix3DTx   = 12
	matrix3DTy   = 13
	matrix3DTz   = 14
	perspectiveW = 11
)

// mat2 is a two-dimensional affine matrix in CSS order:
//
//	| a c e |
//	| b d f |
//	| 0 0 1 |
//
// A point is a column vector, so m.mul(n) applies n first and m second.
type mat2 struct {
	a, b, c, d, e, f float64
}

func identity2D() mat2 { return mat2{a: 1, b: 0, c: 0, d: 1, e: 0, f: 0} }

func (m mat2) mul(n mat2) mat2 {
	return mat2{
		a: m.a*n.a + m.c*n.b,
		b: m.b*n.a + m.d*n.b,
		c: m.a*n.c + m.c*n.d,
		d: m.b*n.c + m.d*n.d,
		e: m.a*n.e + m.c*n.f + m.e,
		f: m.b*n.e + m.d*n.f + m.f,
	}
}

func translate2D(x, y float64) mat2 {
	return mat2{a: 1, b: 0, c: 0, d: 1, e: x, f: y}
}

func scale2D(x, y float64) mat2 {
	return mat2{a: x, b: 0, c: 0, d: y, e: 0, f: 0}
}

func rotate2D(angle float64) mat2 {
	s, c := math.Sincos(angle)
	return mat2{a: c, b: s, c: -s, d: c, e: 0, f: 0}
}

func skewX2D(angle float64) mat2 {
	return mat2{a: 1, b: 0, c: math.Tan(angle), d: 1, e: 0, f: 0}
}

// matrix2D maps one two-dimensional function to its matrix. The bool is
// false for a function that only exists in the 3D list, which compose2D never
// receives.
func matrix2D(f animation.TransformFunction) (mat2, bool) {
	switch f.Name {
	case animation.Translate:
		return translate2D(f.Args[0], f.Args[1]), true
	case animation.TranslateX:
		return translate2D(f.Args[0], 0), true
	case animation.TranslateY:
		return translate2D(0, f.Args[0]), true
	case animation.Scale:
		return scale2D(f.Args[0], f.Args[1]), true
	case animation.ScaleX:
		return scale2D(f.Args[0], 1), true
	case animation.ScaleY:
		return scale2D(1, f.Args[0]), true
	case animation.Rotate, animation.RotateZ:
		return rotate2D(f.Args[0]), true
	case animation.Skew:
		return mat2{a: 1, b: math.Tan(f.Args[1]), c: math.Tan(f.Args[0]), d: 1}, true
	case animation.SkewX:
		return skewX2D(f.Args[0]), true
	case animation.SkewY:
		return mat2{a: 1, b: math.Tan(f.Args[0]), c: 0, d: 1}, true
	case animation.Matrix:
		return mat2{a: f.Args[0], b: f.Args[1], c: f.Args[2], d: f.Args[3], e: f.Args[4], f: f.Args[5]}, true
	}
	return identity2D(), false
}

func compose2D(functions []animation.TransformFunction) mat2 {
	m := identity2D()
	for _, f := range functions {
		next, ok := matrix2D(f)
		if ok {
			m = m.mul(next)
		}
	}
	return m
}

// decompose2D is the two-dimensional unmatrix decomposition: translation,
// scale, rotation, and one XY shear. ScaleX is signed positive and ScaleY
// carries the determinant sign, so the recomposition order
// translate * rotate * skewX * scale reproduces any invertible matrix.
func decompose2D(m mat2) animation.Transform {
	row0x, row0y := m.a, m.b
	scaleX := math.Hypot(row0x, row0y)
	if scaleX != 0 {
		row0x, row0y = row0x/scaleX, row0y/scaleX
	} else {
		row0x, row0y = 1, 0
	}
	// Project the second column onto the rotated axes. The perpendicular
	// component is the signed Y scale; the parallel component is the shear.
	scaleY := -row0y*m.c + row0x*m.d
	skew := 0.0
	if scaleY != 0 {
		skew = (row0x*m.c + row0y*m.d) / scaleY
	}
	return animation.Transform{
		Tx: m.e, Ty: m.f, Sx: scaleX, Sy: scaleY,
		Rotate: math.Atan2(row0y, row0x), SkewX: math.Atan(skew), Set: true,
	}
}

// recompose2D turns a decomposed transform back into its matrix. The inverse
// of decompose2D. Tests use it to prove the round trip.
func recompose2D(t animation.Transform) mat2 {
	return translate2D(t.Tx, t.Ty).
		mul(rotate2D(t.Rotate)).
		mul(skewX2D(t.SkewX)).
		mul(scale2D(t.Sx, t.Sy))
}

// mat4 is a four-by-four matrix in CSS matrix3d column-major order. Element
// (row, col) is m[col*4+row].
type mat4 [16]float64

func identity4() mat4 {
	return mat4{1, 0, 0, 0, 0, 1, 0, 0, 0, 0, 1, 0, 0, 0, 0, 1}
}

func (m mat4) mul(n mat4) mat4 {
	var out mat4
	for col := 0; col < 4; col++ {
		for row := 0; row < 4; row++ {
			sum := 0.0
			for k := 0; k < 4; k++ {
				sum += m[k*4+row] * n[col*4+k]
			}
			out[col*4+row] = sum
		}
	}
	return out
}

// compose3D folds a transform list into one matrix so its two-dimensional
// translation can stay readable in X and Y. The full matrix is not applied by
// phase 02; the engine restamp takes the list instead.
func compose3D(functions []animation.TransformFunction) mat4 {
	m := identity4()
	for _, f := range functions {
		m = m.mul(matrix4(f))
	}
	return m
}

func matrix4(f animation.TransformFunction) mat4 {
	switch f.Name {
	case animation.Translate3D:
		m := identity4()
		m[matrix3DTx], m[matrix3DTy], m[matrix3DTz] = f.Args[0], f.Args[1], f.Args[2]
		return m
	case animation.TranslateZ:
		m := identity4()
		m[matrix3DTz] = f.Args[0]
		return m
	case animation.Scale3D:
		m := identity4()
		m[0], m[5], m[10] = f.Args[0], f.Args[1], f.Args[2]
		return m
	case animation.ScaleZ:
		m := identity4()
		m[10] = f.Args[0]
		return m
	case animation.RotateX:
		return rotateX4(f.Args[0])
	case animation.RotateY:
		return rotateY4(f.Args[0])
	case animation.Rotate3D:
		return rotate3D4(f.Args)
	case animation.Matrix3D:
		var m mat4
		copy(m[:], f.Args)
		return m
	case animation.Perspective:
		m := identity4()
		if f.Args[0] != 0 {
			m[perspectiveW] = -1 / f.Args[0]
		}
		return m
	default:
		m2, ok := matrix2D(f)
		if !ok {
			return identity4()
		}
		return mat4{m2.a, m2.b, 0, 0, m2.c, m2.d, 0, 0, 0, 0, 1, 0, m2.e, m2.f, 0, 1}
	}
}

func rotateX4(angle float64) mat4 {
	s, c := math.Sincos(angle)
	return mat4{1, 0, 0, 0, 0, c, s, 0, 0, -s, c, 0, 0, 0, 0, 1}
}

func rotateY4(angle float64) mat4 {
	s, c := math.Sincos(angle)
	return mat4{c, 0, -s, 0, 0, 1, 0, 0, s, 0, c, 0, 0, 0, 0, 1}
}

func rotate3D4(args []float64) mat4 {
	x, y, z := args[0], args[1], args[2]
	length := math.Sqrt(x*x + y*y + z*z)
	if length == 0 {
		return identity4()
	}
	x, y, z = x/length, y/length, z/length
	s, c := math.Sincos(args[3])
	one := 1 - c
	return mat4{
		c + one*x*x, one*x*y + s*z, one*x*z - s*y, 0,
		one*x*y - s*z, c + one*y*y, one*y*z + s*x, 0,
		one*x*z + s*y, one*y*z - s*x, c + one*z*z, 0,
		0, 0, 0, 1,
	}
}
