package animation

import (
	"math"
	"strconv"
)

// Canonical serialization for the engine re-projection pass.
//
// A keyframe that carries a 3D transform function, and any keyframe on an
// element that already declares a 3D transform, cannot be written through a 2D
// operation matrix. The engine re-projects the element's subtree from a CSS
// transform list instead, so the sampled value has to go back to text. The
// format is the one the cascade's parser reads: the functions in list order,
// every length resolved to px, every angle resolved to deg, and each number in
// its shortest round-trip form so a parse returns the sampled value.

const (
	// pointsPerPixel is layout.Display.PointsPerPixel: one display point is
	// 72/96 of a CSS pixel. TransformFunction stores lengths in points and the
	// engine's parser multiplies a px value by this same factor, so a value
	// written here parses back to the sampled point length.
	pointsPerPixel = 72.0 / 96.0
	// radToDeg writes an interpolated angle, which is in radians, as CSS deg.
	radToDeg = 180 / math.Pi
	// matrixFirstLength is the first argument of matrix() whose unit is a
	// length: e and f. a, b, c, and d are unitless.
	matrixFirstLength = 4
	// matrix3DFirstLength and matrix3DLastLength bound the translations of
	// matrix3d(): m41, m42, and m43 are lengths, m44 is unitless.
	matrix3DFirstLength = 12
	matrix3DLastLength  = 14
	// floatBits is the float64 bit size strconv formats.
	floatBits = 64
)

// TransformCSS returns the canonical CSS transform list for v. A 2D value
// serializes as the matrix() its decomposition describes; a 3D value
// serializes its parsed function list, so rotateY, translateZ, matrix3d, and a
// perspective() function in the list survive the round trip. The result parses
// through the engine's transform parser.
func TransformCSS(v TransformValue) string {
	return string(AppendTransformCSS(nil, v))
}

// AppendTransformCSS appends the canonical CSS transform list for v to dst.
// Pass a reusable buffer to keep the per-frame path allocation-free.
func AppendTransformCSS(dst []byte, v TransformValue) []byte {
	if v.ThreeD && len(v.Functions) > 0 {
		return appendTransformFunctions(dst, v.Functions)
	}

	return appendTransformMatrix(dst, v.Decomposed)
}

// appendTransformMatrix writes the two-dimensional decomposition as matrix().
// The decomposition folds every shear into SkewX, and matrixFromTransform
// applies the CSS product T * R * SkewX * S in that order.
func appendTransformMatrix(dst []byte, t Transform) []byte {
	m := matrixFromTransform(t)
	dst = append(dst, "matrix("...)
	dst = appendNumber(dst, m.A)
	dst = append(dst, ", "...)
	dst = appendNumber(dst, m.B)
	dst = append(dst, ", "...)
	dst = appendNumber(dst, m.C)
	dst = append(dst, ", "...)
	dst = appendNumber(dst, m.D)
	dst = append(dst, ", "...)
	dst = appendLength(dst, m.E)
	dst = append(dst, ", "...)
	dst = appendLength(dst, m.F)

	return append(dst, ')')
}

// appendTransformFunctions writes one parsed list in source order.
func appendTransformFunctions(dst []byte, functions []TransformFunction) []byte {
	for i := range functions {
		if i > 0 {
			dst = append(dst, ' ')
		}
		dst = appendTransformFunction(dst, functions[i])
	}

	return dst
}

// appendTransformFunction writes one function and its arguments.
func appendTransformFunction(dst []byte, f TransformFunction) []byte {
	dst = append(dst, string(f.Name)...)
	dst = append(dst, '(')
	for i := range f.Args {
		if i > 0 {
			dst = append(dst, ", "...)
		}
		dst = appendTransformArg(dst, f.Name, i, len(f.Args), f.Args[i])
	}

	return append(dst, ')')
}

// appendTransformArg writes one argument with the unit its function uses. The
// parser normalizes the optional arguments (translate keeps two, scale keeps
// two), so a canonical value always writes every stored argument.
func appendTransformArg(dst []byte, name TransformFunc, index, count int, value float64) []byte {
	switch name {
	case Translate, TranslateX, TranslateY, Translate3D, TranslateZ, Perspective:
		return appendLength(dst, value)
	case Rotate, RotateX, RotateY, RotateZ, Skew, SkewX, SkewY:
		return appendAngle(dst, value)
	case Rotate3D:
		// The axis is unitless; the last argument is the angle.
		if index == count-1 {
			return appendAngle(dst, value)
		}

		return appendNumber(dst, value)
	case Matrix:
		if index >= matrixFirstLength {
			return appendLength(dst, value)
		}

		return appendNumber(dst, value)
	case Matrix3D:
		if index >= matrix3DFirstLength && index <= matrix3DLastLength {
			return appendLength(dst, value)
		}

		return appendNumber(dst, value)
	case Scale, ScaleX, ScaleY, Scale3D, ScaleZ:
		return appendNumber(dst, value)
	}

	return appendNumber(dst, value)
}

// appendNumber writes a unitless value in shortest round-trip form.
func appendNumber(dst []byte, value float64) []byte {
	return strconv.AppendFloat(dst, value, 'f', -1, floatBits)
}

// appendLength writes a length stored in display points as CSS pixels.
func appendLength(dst []byte, points float64) []byte {
	dst = strconv.AppendFloat(dst, points/pointsPerPixel, 'f', -1, floatBits)

	return append(dst, "px"...)
}

// appendAngle writes an angle stored in radians as CSS degrees.
func appendAngle(dst []byte, radians float64) []byte {
	dst = strconv.AppendFloat(dst, radians*radToDeg, 'f', -1, floatBits)

	return append(dst, "deg"...)
}
