package animation

import (
	"errors"
	"math"
	"testing"
)

// matrixTolerance4 compares two matrices built through different composition
// paths.
const matrixTolerance4 = 1e-9

// normalized4 returns the same projective matrix with m44 set to 1, the form
// the decomposition works in.
func normalized4(m Matrix4) Matrix4 {
	if m[indexM44] == 0 || m[indexM44] == 1 {
		return m
	}
	for i := range m {
		m[i] /= m[indexM44]
	}
	return m
}

// requireMatrix4Near compares two sampled matrices element by element.
func requireMatrix4Near(t *testing.T, got, want Matrix4, tolerance float64) {
	t.Helper()
	for i := range got {
		if math.Abs(got[i]-want[i]) > tolerance {
			t.Fatalf("matrix element %d = %v, want %v", i, got[i], want[i])
		}
	}
}

// matrixArgs copies a matrix into the argument list of matrix3d().
func matrixArgs(m Matrix4) []float64 {
	args := make([]float64, len(m))
	copy(args, m[:])
	return args
}

func identityMatrix3D() []float64 {
	return matrixArgs(IdentityMatrix4())
}

func TestDecompose3DRecomposeRoundTrip(t *testing.T) {
	fixtures := []struct {
		name      string
		functions []TransformFunction
	}{
		{"identity", nil},
		{"translate3d", []TransformFunction{{Name: Translate3D, Args: []float64{10, 20, 30}}}},
		{"scale3d", []TransformFunction{{Name: Scale3D, Args: []float64{2, 3, 4}}}},
		{"rotateX", []TransformFunction{{Name: RotateX, Args: []float64{0.7}}}},
		{"rotateY", []TransformFunction{{Name: RotateY, Args: []float64{-0.4}}}},
		{"rotateZ", []TransformFunction{{Name: RotateZ, Args: []float64{1.2}}}},
		{"rotate3d", []TransformFunction{{Name: Rotate3D, Args: []float64{1, 2, 3, 1.1}}}},
		{"translate rotate scale", []TransformFunction{
			{Name: Translate3D, Args: []float64{4, 5, 6}},
			{Name: RotateY, Args: []float64{0.5}},
			{Name: Scale3D, Args: []float64{2, 3, 4}},
		}},
		{"perspective", []TransformFunction{
			{Name: Perspective, Args: []float64{200}},
			{Name: RotateY, Args: []float64{math.Pi / 4}},
			{Name: TranslateZ, Args: []float64{10}},
		}},
		{"skew and 2D scale", []TransformFunction{
			{Name: SkewX, Args: []float64{0.3}},
			{Name: Scale, Args: []float64{2, 3}},
		}},
		{"reflection", []TransformFunction{{Name: Scale, Args: []float64{-1, 1}}}},
	}

	for _, fixture := range fixtures {
		t.Run(fixture.name, func(t *testing.T) {
			want := Matrix4FromFunctions(fixture.functions)
			decomposed, err := Decompose3D(want)
			if err != nil {
				t.Fatal(err)
			}
			// Decompose3D normalizes by m44, so the round trip returns the
			// projectively equivalent matrix whose m44 is 1.
			requireMatrix4Near(t, Compose3D(decomposed), normalized4(want), matrixTolerance4)
		})
	}
}

func TestDecompose3DComponents(t *testing.T) {
	translate, err := Decompose3D(Matrix4FromFunctions(
		[]TransformFunction{{Name: Translate3D, Args: []float64{10, 20, 30}}}))
	if err != nil {
		t.Fatal(err)
	}
	if translate.Translate != [3]float64{10, 20, 30} {
		t.Fatalf("translation = %v, want [10 20 30]", translate.Translate)
	}
	if translate.Scale != [3]float64{1, 1, 1} || translate.Perspective != [4]float64{0, 0, 0, 1} {
		t.Fatalf("translation decomposition = %+v, want identity scale and perspective", translate)
	}

	perspective, err := Decompose3D(Matrix4FromFunctions(
		[]TransformFunction{{Name: Perspective, Args: []float64{200}}}))
	if err != nil {
		t.Fatal(err)
	}
	if math.Abs(perspective.Perspective[2]-(-1.0/200)) > matrixTolerance4 ||
		perspective.Perspective[3] != 1 {
		t.Fatalf("perspective = %v, want the m34 divisor -1/200", perspective.Perspective)
	}

	rotation, err := Decompose3D(Matrix4FromFunctions(
		[]TransformFunction{{Name: RotateY, Args: []float64{math.Pi / 2}}}))
	if err != nil {
		t.Fatal(err)
	}
	// The quaternion magnitude is half the angle; its sign is the convention
	// Compose3D expects, not an authored value.
	if math.Abs(math.Abs(rotation.Quaternion[1])-math.Sin(math.Pi/4)) > matrixTolerance4 ||
		math.Abs(rotation.Quaternion[3]-math.Cos(math.Pi/4)) > matrixTolerance4 {
		t.Fatalf("quaternion = %v, want a 90 degree rotation about Y", rotation.Quaternion)
	}
}

// TestInterpolateRotateXZeroTo180 checks the spec's 180 degree example. The
// lists match, so CSS interpolates component-wise and the midpoint is
// rotateX(90); the matrix fallback is covered separately below.
func TestInterpolateRotateXZeroTo180(t *testing.T) {
	value, err := interpolateTransform(
		rotateValue(RotateX, 0, true), rotateValue(RotateX, math.Pi, true), .5)
	if err != nil {
		t.Fatal(err)
	}
	if !value.ThreeD || len(value.Functions) != 1 || value.Functions[0].Name != RotateX {
		t.Fatalf("matching 3D pair sampled %+v, want rotateX", value.Functions)
	}
	if math.Abs(value.Functions[0].Args[0]-math.Pi/2) > matrixTolerance4 {
		t.Fatalf("rotateX midpoint = %v, want 90 degrees", value.Functions[0].Args[0])
	}
}

// TestInterpolateMatrix3DHalfTurn covers a half turn through the matrix path.
// A 180 degree rotation has two equally short quaternion paths, so the sign of
// the midpoint is the quaternion convention's; either way it is a quarter turn
// about X.
func TestInterpolateMatrix3DHalfTurn(t *testing.T) {
	rotation := rotateX4(math.Pi)
	to := TransformValue{
		Decomposed: IdentityTransform(),
		Functions:  []TransformFunction{{Name: Matrix3D, Args: matrixArgs(rotation)}},
		ThreeD:     true,
	}
	value, err := interpolateTransform(TransformValue{Decomposed: IdentityTransform(), ThreeD: true,
		Functions: []TransformFunction{{Name: Matrix3D, Args: identityMatrix3D()}}}, to, .5)
	if err != nil {
		t.Fatal(err)
	}
	m := Matrix4FromFunctions(value.Functions)
	if math.Abs(m[5]) > matrixTolerance4 || math.Abs(math.Abs(m[6])-1) > matrixTolerance4 {
		t.Fatalf("half-turn midpoint = %v, want a quarter turn about X", m)
	}
}

// TestInterpolateRotateYZeroTo90 checks the spec's 90 degree example inside a
// mismatched list, which is the case the bind used to reject.
func TestInterpolateRotateYZeroTo90(t *testing.T) {
	from := TransformValue{Decomposed: IdentityTransform(), ThreeD: true,
		Functions: []TransformFunction{{Name: RotateY, Args: []float64{0}}}}
	to := TransformValue{Decomposed: IdentityTransform(), ThreeD: true,
		Functions: []TransformFunction{
			{Name: TranslateZ, Args: []float64{0}},
			{Name: RotateY, Args: []float64{math.Pi / 2}},
		}}
	value, err := interpolateTransform(from, to, .5)
	if err != nil {
		t.Fatal(err)
	}
	requireMatrix4Near(t, Matrix4FromFunctions(value.Functions), rotateY4(math.Pi/4), matrixTolerance4)
}

// TestInterpolatePerspectiveBearingMatrix covers the projective pair: both
// perspectives decompose, the rotation slerps, and the divisor interpolates.
func TestInterpolatePerspectiveBearingMatrix(t *testing.T) {
	from := TransformValue{Decomposed: IdentityTransform(), ThreeD: true,
		Functions: []TransformFunction{
			{Name: Perspective, Args: []float64{200}},
			{Name: RotateZ, Args: []float64{0}},
		}}
	to := TransformValue{Decomposed: IdentityTransform(), ThreeD: true,
		Functions: []TransformFunction{
			{Name: Perspective, Args: []float64{400}},
			{Name: RotateZ, Args: []float64{math.Pi / 2}},
		}}
	value, err := interpolateTransform(from, to, .5)
	if err != nil {
		t.Fatal(err)
	}
	m := Matrix4FromFunctions(value.Functions)
	// The perspective divisors -1/200 and -1/400 average to -3/800, and a Z
	// rotation leaves the projective row alone.
	if math.Abs(m[indexM34]-(-3.0/800)) > matrixTolerance4 {
		t.Fatalf("m34 = %v, want the averaged perspective divisor %v", m[indexM34], -3.0/800)
	}
	if !value.ThreeD {
		t.Fatal("a perspective-bearing sample dropped the restamp flag")
	}
}

// TestInterpolateTwoDimensionalListsStayTwoDimensional pins the fast path:
// neither endpoint carries a 3D function, so the pair must not be decomposed
// through the four-by-four algorithm.
func TestInterpolateTwoDimensionalListsStayTwoDimensional(t *testing.T) {
	rotate := TransformValue{
		Decomposed: Transform{Sx: 1, Sy: 1, Rotate: math.Pi / 2, Set: true},
		Functions:  []TransformFunction{{Name: Rotate, Args: []float64{math.Pi / 2}}},
	}
	value, err := interpolateTransform(translateValue(10, 0), rotate, .5)
	if err != nil {
		t.Fatal(err)
	}
	if value.ThreeD || value.Functions != nil {
		t.Fatalf("2D pair took the 3D path: %+v", value)
	}
	if value.Decomposed.Tx != 5 || math.Abs(value.Decomposed.Rotate-math.Pi/4) > matrixTolerance4 {
		t.Fatalf("decomposed = %+v, want the 2D midpoint", value.Decomposed)
	}
}

// TestSampleTransformStepsDiscretelyForASingularMatrix pins the CSS fallback:
// a non-invertible endpoint cannot be decomposed, so the interval picks an
// endpoint at its midpoint instead of snapping the whole animation.
func TestSampleTransformFallsBackToDiscreteForASingularMatrix(t *testing.T) {
	from := rotateValue(RotateY, 0, true)
	to := TransformValue{Decomposed: IdentityTransform(), ThreeD: true,
		Functions: []TransformFunction{{Name: ScaleZ, Args: []float64{0}}}}
	effect := Effect{Frames: []Keyframe{{Offset: 0, Transform: from}, {Offset: 1, Transform: to}}}

	before, ok := effect.sampleTransform(.4)
	if !ok || before.Functions[0].Name != RotateY {
		t.Fatalf("sample at 40%% = %+v, want the first endpoint", before)
	}
	after, ok := effect.sampleTransform(.6)
	if !ok || after.Functions[0].Name != ScaleZ {
		t.Fatalf("sample at 60%% = %+v, want the last endpoint", after)
	}
}

func TestDecompose3DSingularMatrixError(t *testing.T) {
	for _, m := range []Matrix4{
		{},
		scale4(0, 1, 1),
		scale4(1, 1, 0),
	} {
		if _, err := Decompose3D(m); !errors.Is(err, ErrSingularMatrix) {
			t.Fatalf("Decompose3D(%v) = %v, want ErrSingularMatrix", m, err)
		}
	}
}
