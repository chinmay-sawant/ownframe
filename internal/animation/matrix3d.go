package animation

import "math"

// Four-by-four transform matrices.
//
// A Matrix4 is stored in CSS matrix3d order, which is column-major: the flat
// index of matrix element M(row, col), with both one-based, is
// (col-1)*4 + (row-1). So M41, the translation x, is the thirteenth value and
// M34, the perspective divisor, is the twelfth. A point is a column vector, so
// Mul applies its right operand first.

// Matrix element indexes the decomposition reads and writes. The names use
// CSS matrix notation, where M14 is column 1, row 4 (perspective x) and M41 is
// column 4, row 1 (translation x).
const (
	indexM14 = 3
	indexM24 = 7
	indexM34 = 11
	indexM44 = 15
	indexM41 = 12
	indexM42 = 13
	indexM43 = 14
)

// Matrix4 is a four-by-four homogeneous transform matrix in CSS matrix3d order.
type Matrix4 [16]float64

// IdentityMatrix4 returns the neutral Matrix4.
func IdentityMatrix4() Matrix4 {
	return Matrix4{1, 0, 0, 0, 0, 1, 0, 0, 0, 0, 1, 0, 0, 0, 0, 1}
}

// Mul returns m * n, which applies n first and m second.
func (m Matrix4) Mul(n Matrix4) Matrix4 {
	var out Matrix4
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

// mulVec returns m * v for a column vector v.
func (m Matrix4) mulVec(v [4]float64) [4]float64 {
	var out [4]float64
	for row := 0; row < 4; row++ {
		sum := 0.0
		for col := 0; col < 4; col++ {
			sum += m[col*4+row] * v[col]
		}
		out[row] = sum
	}
	return out
}

// finite reports whether every element is a finite number. A non-finite matrix
// cannot be decomposed; treating it as singular keeps a bad authored value from
// poisoning a sampled frame.
func (m Matrix4) finite() bool {
	for _, v := range m {
		if math.IsNaN(v) || math.IsInf(v, 0) {
			return false
		}
	}
	return true
}

// Matrix4FromFunctions folds a transform list into one matrix in source order,
// so the first function is applied last. It is the four-by-four counterpart of
// the two-dimensional composition in internal/cssanimation.
func Matrix4FromFunctions(functions []TransformFunction) Matrix4 {
	m := IdentityMatrix4()
	for _, f := range functions {
		m = m.Mul(matrix4For(f))
	}
	return m
}

// matrix4For maps one transform function to its four-by-four matrix. Lengths
// arrive in display points and angles in radians, the units the parser stores.
func matrix4For(f TransformFunction) Matrix4 {
	switch f.Name {
	case Translate:
		return translation4(f.Args[0], f.Args[1], 0)
	case TranslateX:
		return translation4(f.Args[0], 0, 0)
	case TranslateY:
		return translation4(0, f.Args[0], 0)
	case Translate3D:
		return translation4(f.Args[0], f.Args[1], f.Args[2])
	case TranslateZ:
		return translation4(0, 0, f.Args[0])
	case Scale:
		return scale4(f.Args[0], f.Args[1], 1)
	case ScaleX:
		return scale4(f.Args[0], 1, 1)
	case ScaleY:
		return scale4(1, f.Args[0], 1)
	case Scale3D:
		return scale4(f.Args[0], f.Args[1], f.Args[2])
	case ScaleZ:
		return scale4(1, 1, f.Args[0])
	case Rotate, RotateZ:
		return rotateZ4(f.Args[0])
	case RotateX:
		return rotateX4(f.Args[0])
	case RotateY:
		return rotateY4(f.Args[0])
	case Rotate3D:
		return rotate3D4(f.Args)
	case Skew:
		return skew4(f.Args[0], f.Args[1])
	case SkewX:
		return skew4(f.Args[0], 0)
	case SkewY:
		return skew4(0, f.Args[0])
	case Matrix:
		return Matrix4{
			f.Args[0], f.Args[1], 0, 0,
			f.Args[2], f.Args[3], 0, 0,
			0, 0, 1, 0,
			f.Args[4], f.Args[5], 0, 1,
		}
	case Matrix3D:
		var m Matrix4
		copy(m[:], f.Args)
		return m
	case Perspective:
		m := IdentityMatrix4()
		if f.Args[0] != 0 {
			m[indexM34] = -1 / f.Args[0]
		}
		return m
	}
	return IdentityMatrix4()
}

func translation4(x, y, z float64) Matrix4 {
	m := IdentityMatrix4()
	m[indexM41], m[indexM42], m[indexM43] = x, y, z
	return m
}

func scale4(x, y, z float64) Matrix4 {
	return Matrix4{x, 0, 0, 0, 0, y, 0, 0, 0, 0, z, 0, 0, 0, 0, 1}
}

// skew4 builds the two-dimensional skew matrix: skewX shears x by y and skewY
// shears y by x.
func skew4(x, y float64) Matrix4 {
	return Matrix4{1, math.Tan(y), 0, 0, math.Tan(x), 1, 0, 0, 0, 0, 1, 0, 0, 0, 0, 1}
}

func rotateX4(angle float64) Matrix4 {
	s, c := math.Sincos(angle)
	return Matrix4{1, 0, 0, 0, 0, c, s, 0, 0, -s, c, 0, 0, 0, 0, 1}
}

func rotateY4(angle float64) Matrix4 {
	s, c := math.Sincos(angle)
	return Matrix4{c, 0, -s, 0, 0, 1, 0, 0, s, 0, c, 0, 0, 0, 0, 1}
}

func rotateZ4(angle float64) Matrix4 {
	s, c := math.Sincos(angle)
	return Matrix4{c, s, 0, 0, -s, c, 0, 0, 0, 0, 1, 0, 0, 0, 0, 1}
}

func rotate3D4(args []float64) Matrix4 {
	x, y, z := args[0], args[1], args[2]
	length := math.Sqrt(x*x + y*y + z*z)
	if length == 0 {
		return IdentityMatrix4()
	}
	x, y, z = x/length, y/length, z/length
	s, c := math.Sincos(args[3])
	one := 1 - c
	return Matrix4{
		c + one*x*x, one*x*y + s*z, one*x*z - s*y, 0,
		one*x*y - s*z, c + one*y*y, one*y*z + s*x, 0,
		one*x*z + s*y, one*y*z - s*x, c + one*z*z, 0,
		0, 0, 0, 1,
	}
}

// Decomposition3D is a four-by-four matrix split into the components the CSS
// Transforms Level 2 algorithm interpolates. Translation is in display points,
// scale is a multiplier, skew is three shear factors (XY, XZ, YZ), perspective
// is the fourth row of the matrix, and Quaternion is (x, y, z, w).
type Decomposition3D struct {
	Translate   [3]float64
	Scale       [3]float64
	Skew        [3]float64
	Perspective [4]float64
	Quaternion  [4]float64
}

// Decompose3D splits m into translation, scale, skew, perspective, and a
// rotation quaternion, following the CSS Transforms Level 2 unmatrix algorithm.
//
// The quaternion extraction and the matching rotation matrix in Compose3D use
// the WebKit/Blink convention rather than the literal pseudocode. The Level 2
// pseudocode's sign fix-ups and its recomposition matrix are transposed
// relative to each other (csswg-drafts issue 3230), so a literal port does not
// round-trip a rotation; the browser pair does, and it is what the spec's
// examples intend.
//
// A singular matrix, one whose m44 is zero or whose upper three-by-three part
// cannot be inverted, reports ErrSingularMatrix.
func Decompose3D(m Matrix4) (Decomposition3D, error) {
	if m[indexM44] == 0 || !m.finite() {
		return Decomposition3D{}, ErrSingularMatrix
	}
	if m[indexM44] != 1 {
		for i := range m {
			m[i] /= m[indexM44]
		}
	}

	perspectiveMatrix := m
	perspectiveMatrix[indexM14], perspectiveMatrix[indexM24], perspectiveMatrix[indexM34] = 0, 0, 0
	perspectiveMatrix[indexM44] = 1
	inversePerspective, ok := inverse4(perspectiveMatrix)
	if !ok {
		return Decomposition3D{}, ErrSingularMatrix
	}

	var d Decomposition3D
	if m[indexM14] != 0 || m[indexM24] != 0 || m[indexM34] != 0 {
		rightHandSide := [4]float64{m[indexM14], m[indexM24], m[indexM34], m[indexM44]}
		// The perspective vector satisfies rhs = perspective * A, so it is
		// the transposed inverse times rhs: the spec's row-vector multiply.
		d.Perspective = transpose4(inversePerspective).mulVec(rightHandSide)
		m[indexM14], m[indexM24], m[indexM34], m[indexM44] = 0, 0, 0, 1
	} else {
		d.Perspective = [4]float64{0, 0, 0, 1}
	}

	d.Translate = [3]float64{m[indexM41], m[indexM42], m[indexM43]}
	m[indexM41], m[indexM42], m[indexM43] = 0, 0, 0

	// The three columns of the upper three-by-three part hold the basis
	// vectors. Gram-Schmidt recovers scale and shear.
	var row [3][3]float64
	for col := 0; col < 3; col++ {
		copy(row[col][:], m[col*4:col*4+3])
	}

	d.Scale[0] = length3(row[0])
	row[0] = normalize3(row[0])

	d.Skew[0] = dot3(row[0], row[1])
	row[1] = combine3(row[1], row[0], 1, -d.Skew[0])

	d.Scale[1] = length3(row[1])
	row[1] = normalize3(row[1])
	d.Skew[0] /= d.Scale[1]

	d.Skew[1] = dot3(row[0], row[2])
	row[2] = combine3(row[2], row[0], 1, -d.Skew[1])
	d.Skew[2] = dot3(row[1], row[2])
	row[2] = combine3(row[2], row[1], 1, -d.Skew[2])

	d.Scale[2] = length3(row[2])
	row[2] = normalize3(row[2])
	d.Skew[1] /= d.Scale[2]
	d.Skew[2] /= d.Scale[2]

	if dot3(row[0], cross3(row[1], row[2])) < 0 {
		for i := 0; i < 3; i++ {
			d.Scale[i] *= -1
			row[i] = negate3(row[i])
		}
	}

	d.Quaternion = quaternionFromRows(row)

	return d, nil
}

// Compose3D rebuilds a four-by-four matrix from a decomposition. It is the
// inverse of Decompose3D: perspective, then translation, rotation, the three
// shears, and scale, all post-multiplied in that order.
func Compose3D(d Decomposition3D) Matrix4 {
	m := IdentityMatrix4()
	for i := 0; i < 4; i++ {
		m[i*4+3] = d.Perspective[i]
	}
	for i := 0; i < 4; i++ {
		for j := 0; j < 3; j++ {
			m[3*4+i] += d.Translate[j] * m[j*4+i]
		}
	}

	m = m.Mul(quaternionMatrix(d.Quaternion))

	temp := IdentityMatrix4()
	if d.Skew[2] != 0 {
		temp[2*4+1] = d.Skew[2]
		m = m.Mul(temp)
	}
	if d.Skew[1] != 0 {
		temp[2*4+1] = 0
		temp[2*4+0] = d.Skew[1]
		m = m.Mul(temp)
	}
	if d.Skew[0] != 0 {
		temp[2*4+0] = 0
		temp[1*4+0] = d.Skew[0]
		m = m.Mul(temp)
	}

	for col := 0; col < 3; col++ {
		for row := 0; row < 4; row++ {
			m[col*4+row] *= d.Scale[col]
		}
	}

	return m
}

// quaternionMatrix builds the rotation matrix for a unit quaternion
// (x, y, z, w).
func quaternionMatrix(q [4]float64) Matrix4 {
	x, y, z, w := q[0], q[1], q[2], q[3]
	xx, yy, zz := x*x, y*y, z*z
	return Matrix4{
		1 - 2*(yy+zz), 2 * (x*y - z*w), 2 * (x*z + y*w), 0,
		2 * (x*y + z*w), 1 - 2*(xx+zz), 2 * (y*z - x*w), 0,
		2 * (x*z - y*w), 2 * (y*z + x*w), 1 - 2*(xx+yy), 0,
		0, 0, 0, 1,
	}
}

// quaternionTrace is the threshold below which the trace formula loses
// precision and the largest-diagonal branch takes over.
const quaternionTrace = 1e-4

// Quaternion extraction factors. The trace branch divides the trace by two and
// the largest-diagonal branch doubles the square root.
const (
	quaternionHalf    = 0.5
	quaternionQuarter = 0.25
	quaternionDouble  = 2.0
)

// quaternionFromRows extracts the rotation quaternion from the orthonormal
// basis rows. The branch on the largest diagonal keeps a rotation near 180
// degrees numerically stable.
func quaternionFromRows(row [3][3]float64) [4]float64 {
	var q [4]float64
	t := row[0][0] + row[1][1] + row[2][2] + 1
	switch {
	case t > quaternionTrace:
		s := quaternionHalf / math.Sqrt(t)
		q[3] = quaternionQuarter / s
		q[0] = (row[2][1] - row[1][2]) * s
		q[1] = (row[0][2] - row[2][0]) * s
		q[2] = (row[1][0] - row[0][1]) * s
	case row[0][0] > row[1][1] && row[0][0] > row[2][2]:
		s := math.Sqrt(1+row[0][0]-row[1][1]-row[2][2]) * quaternionDouble
		q[0] = quaternionQuarter * s
		q[1] = (row[0][1] + row[1][0]) / s
		q[2] = (row[0][2] + row[2][0]) / s
		q[3] = (row[2][1] - row[1][2]) / s
	case row[1][1] > row[2][2]:
		s := math.Sqrt(1+row[1][1]-row[0][0]-row[2][2]) * quaternionDouble
		q[0] = (row[0][1] + row[1][0]) / s
		q[1] = quaternionQuarter * s
		q[2] = (row[1][2] + row[2][1]) / s
		q[3] = (row[0][2] - row[2][0]) / s
	default:
		s := math.Sqrt(1+row[2][2]-row[0][0]-row[1][1]) * quaternionDouble
		q[0] = (row[0][2] + row[2][0]) / s
		q[1] = (row[1][2] + row[2][1]) / s
		q[2] = quaternionQuarter * s
		q[3] = (row[1][0] - row[0][1]) / s
	}
	return q
}

// interpolateDecomposition blends two decompositions: every component linearly,
// the quaternion by spherical linear interpolation.
func interpolateDecomposition(a, b Decomposition3D, t float64) Decomposition3D {
	var out Decomposition3D
	for i := range out.Translate {
		out.Translate[i] = lerp(a.Translate[i], b.Translate[i], t)
	}
	for i := range out.Scale {
		out.Scale[i] = lerp(a.Scale[i], b.Scale[i], t)
	}
	for i := range out.Skew {
		out.Skew[i] = lerp(a.Skew[i], b.Skew[i], t)
	}
	for i := range out.Perspective {
		out.Perspective[i] = lerp(a.Perspective[i], b.Perspective[i], t)
	}
	out.Quaternion = slerp(a.Quaternion, b.Quaternion, t)
	return out
}

// slerpParallel is the dot-product threshold at which two quaternions are
// parallel enough that the great-circle formula divides by zero.
const slerpParallel = 1e-5

// slerp interpolates two quaternions along the great circle. A zero or
// near-zero angle returns the first quaternion, which is the identity-pair
// case; product is clamped because rounding can push it outside [-1, 1].
func slerp(a, b [4]float64, t float64) [4]float64 {
	product := a[0]*b[0] + a[1]*b[1] + a[2]*b[2] + a[3]*b[3]
	product = math.Min(math.Max(product, -1), 1)
	if math.Abs(product) >= 1-slerpParallel {
		return a
	}
	theta := math.Acos(product)
	w := math.Sin(t*theta) / math.Sqrt(1-product*product)
	scaleA := math.Cos(t*theta) - product*w
	var out [4]float64
	for i := range out {
		out[i] = a[i]*scaleA + b[i]*w
	}
	return out
}

// transpose4 returns the transpose of m.
func transpose4(m Matrix4) Matrix4 {
	var out Matrix4
	for col := 0; col < 4; col++ {
		for row := 0; row < 4; row++ {
			out[col*4+row] = m[row*4+col]
		}
	}
	return out
}

// inverse4 inverts m with Gauss-Jordan elimination and partial pivoting. ok is
// false for a singular matrix, which is the decomposition's failure case.
func inverse4(m Matrix4) (Matrix4, bool) {
	var aug [4][8]float64
	for row := 0; row < 4; row++ {
		for col := 0; col < 4; col++ {
			aug[row][col] = m[col*4+row]
		}
		aug[row][4+row] = 1
	}
	for col := 0; col < 4; col++ {
		pivot := col
		for row := col + 1; row < 4; row++ {
			if math.Abs(aug[row][col]) > math.Abs(aug[pivot][col]) {
				pivot = row
			}
		}
		if aug[pivot][col] == 0 {
			return Matrix4{}, false
		}
		aug[pivot], aug[col] = aug[col], aug[pivot]
		scale := 1 / aug[col][col]
		for j := 0; j < 8; j++ {
			aug[col][j] *= scale
		}
		for row := 0; row < 4; row++ {
			if row == col || aug[row][col] == 0 {
				continue
			}
			factor := aug[row][col]
			for j := 0; j < 8; j++ {
				aug[row][j] -= factor * aug[col][j]
			}
		}
	}
	var out Matrix4
	for row := 0; row < 4; row++ {
		for col := 0; col < 4; col++ {
			out[col*4+row] = aug[row][4+col]
		}
	}
	return out, true
}

func length3(v [3]float64) float64 {
	return math.Sqrt(v[0]*v[0] + v[1]*v[1] + v[2]*v[2])
}

// normalize3 scales v to unit length. A zero vector stays zero, matching the
// reference implementation, though a singular matrix never reaches it.
func normalize3(v [3]float64) [3]float64 {
	length := length3(v)
	if length == 0 {
		return v
	}
	return [3]float64{v[0] / length, v[1] / length, v[2] / length}
}

func combine3(a, b [3]float64, aScale, bScale float64) [3]float64 {
	return [3]float64{
		aScale*a[0] + bScale*b[0],
		aScale*a[1] + bScale*b[1],
		aScale*a[2] + bScale*b[2],
	}
}

func negate3(v [3]float64) [3]float64 {
	return [3]float64{-v[0], -v[1], -v[2]}
}

func dot3(a, b [3]float64) float64 {
	return a[0]*b[0] + a[1]*b[1] + a[2]*b[2]
}

func cross3(a, b [3]float64) [3]float64 {
	return [3]float64{
		a[1]*b[2] - a[2]*b[1],
		a[2]*b[0] - a[0]*b[2],
		a[0]*b[1] - a[1]*b[0],
	}
}
