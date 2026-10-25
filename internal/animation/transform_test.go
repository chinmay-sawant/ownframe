package animation

import (
	"errors"
	"math"
	"strings"
	"testing"
	"time"

	"github.com/chinmay-sawant/blinkless/layout"
	"github.com/chinmay-sawant/ownframe/internal/cssmotion"
)

const matrixTolerance = 1e-9

func translateValue(tx, ty float64) TransformValue {
	return TransformValue{
		Decomposed: Transform{Tx: tx, Ty: ty, Sx: 1, Sy: 1, Set: true},
		Functions:  []TransformFunction{{Name: Translate, Args: []float64{tx, ty}}},
	}
}

func rotateValue(name TransformFunc, angle float64, threeD bool) TransformValue {
	decomposed := IdentityTransform()
	if !threeD {
		decomposed.Rotate = angle
	}
	return TransformValue{
		Decomposed: decomposed,
		Functions:  []TransformFunction{{Name: name, Args: []float64{angle}}},
		ThreeD:     threeD,
	}
}

func TestInterpolateMatchingTwoDimensionalLists(t *testing.T) {
	value, err := interpolateTransform(translateValue(0, 0), translateValue(20, 10), .5)
	if err != nil {
		t.Fatal(err)
	}
	if value.Decomposed.Tx != 10 || value.Decomposed.Ty != 5 {
		t.Fatalf("decomposed = %+v, want translate 10,5", value.Decomposed)
	}
	if len(value.Functions) != 1 || value.Functions[0].Args[0] != 10 || value.Functions[0].Args[1] != 5 {
		t.Fatalf("functions = %+v, want translate 10,5", value.Functions)
	}
}

func TestInterpolateMatchingThreeItemList(t *testing.T) {
	from := TransformValue{
		Decomposed: Transform{Tx: 10, Ty: 20, Sx: 2, Sy: 2, Set: true},
		Functions: []TransformFunction{
			{Name: Translate, Args: []float64{10, 20}},
			{Name: Rotate, Args: []float64{0}},
			{Name: Scale, Args: []float64{2, 2}},
		},
	}
	to := TransformValue{
		Decomposed: Transform{Tx: 30, Ty: 40, Sx: 4, Sy: 3, Set: true},
		Functions: []TransformFunction{
			{Name: Translate, Args: []float64{30, 40}},
			{Name: Rotate, Args: []float64{math.Pi}},
			{Name: Scale, Args: []float64{4, 3}},
		},
	}
	value, err := interpolateTransform(from, to, .5)
	if err != nil {
		t.Fatal(err)
	}
	if len(value.Functions) != 3 {
		t.Fatalf("functions = %d, want 3", len(value.Functions))
	}
	if value.Functions[0].Args[0] != 20 || value.Functions[1].Args[0] != math.Pi/2 ||
		value.Functions[2].Args[1] != 2.5 {
		t.Fatalf("functions = %+v, want component-wise midpoint", value.Functions)
	}
}

func TestInterpolateMismatchedTwoDimensionalLists(t *testing.T) {
	rotate := TransformValue{
		Decomposed: Transform{Sx: 1, Sy: 1, Rotate: math.Pi / 2, Set: true},
		Functions:  []TransformFunction{{Name: Rotate, Args: []float64{math.Pi / 2}}},
	}
	value, err := interpolateTransform(translateValue(10, 0), rotate, .5)
	if err != nil {
		t.Fatal(err)
	}
	if value.Decomposed.Tx != 5 || math.Abs(value.Decomposed.Rotate-math.Pi/4) > 1e-9 {
		t.Fatalf("decomposed = %+v, want midpoint translation and rotation", value.Decomposed)
	}
	if value.ThreeD {
		t.Fatal("a 2D pair took the 3D path")
	}
	if value.Functions != nil {
		t.Fatal("a mismatched list kept a function list that cannot be applied")
	}
}

func TestInterpolateMatchingThreeDimensionalLists(t *testing.T) {
	value, err := interpolateTransform(
		TransformValue{Decomposed: IdentityTransform(), ThreeD: true,
			Functions: []TransformFunction{{Name: RotateY, Args: []float64{0}}}},
		TransformValue{Decomposed: IdentityTransform(), ThreeD: true,
			Functions: []TransformFunction{{Name: RotateY, Args: []float64{math.Pi / 2}}}},
		.5)
	if err != nil {
		t.Fatal(err)
	}
	if !value.ThreeD || len(value.Functions) != 1 || math.Abs(value.Functions[0].Args[0]-math.Pi/4) > 1e-9 {
		t.Fatalf("functions = %+v, want rotateY at 45 degrees", value.Functions)
	}
}

func TestInterpolateThreeDimensionalListWithUnderlyingNone(t *testing.T) {
	to := TransformValue{Decomposed: IdentityTransform(), ThreeD: true,
		Functions: []TransformFunction{{Name: RotateY, Args: []float64{math.Pi / 2}}}}
	value, err := interpolateTransform(TransformValue{Decomposed: IdentityTransform()}, to, .5)
	if err != nil {
		t.Fatal(err)
	}
	if math.Abs(value.Functions[0].Args[0]-math.Pi/4) > 1e-9 {
		t.Fatalf("rotateY = %v, want 45 degrees", value.Functions[0].Args[0])
	}
}

func TestInterpolateMismatchedThreeDimensionalLists(t *testing.T) {
	from := rotateValue(RotateX, 0, true)
	to := rotateValue(RotateY, math.Pi/2, true)
	value, err := interpolateTransform(from, to, .5)
	if err != nil {
		t.Fatal(err)
	}
	if !value.ThreeD || len(value.Functions) != 1 || value.Functions[0].Name != Matrix3D {
		t.Fatalf("mismatched 3D pair sampled %+v, want one matrix3d list", value.Functions)
	}
	// Slerping identity to rotateY(90deg) is rotateY(45deg).
	requireMatrix4Near(t, Matrix4FromFunctions(value.Functions), rotateY4(math.Pi/4), matrixTolerance)
}

func TestInterpolatePerspectiveAgainstNone(t *testing.T) {
	to := TransformValue{
		Decomposed: IdentityTransform(),
		Functions:  []TransformFunction{{Name: Perspective, Args: []float64{100}}},
		ThreeD:     true,
	}
	value, err := interpolateTransform(TransformValue{Decomposed: IdentityTransform()}, to, .5)
	if err != nil {
		t.Fatal(err)
	}
	// The perspective matrix is [0, 0, -1/d, 1] and identity carries
	// [0, 0, 0, 1], so the midpoint is the matrix for d=200.
	m := Matrix4FromFunctions(value.Functions)
	if math.Abs(m[indexM34]-(-1.0/200)) > matrixTolerance {
		t.Fatalf("perspective midpoint m34 = %v, want %v", m[indexM34], -1.0/200)
	}
}

func TestInterpolateMatrix3DPair(t *testing.T) {
	rotation := rotateX4(math.Pi / 2)
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
	// Identity to rotateX(90deg) passes through rotateX(45deg).
	requireMatrix4Near(t, Matrix4FromFunctions(value.Functions), rotateX4(math.Pi/4), matrixTolerance)
}

func TestComposeAppliesTwoDimensionalTranslation(t *testing.T) {
	op := &layout.DisplayOp{Kind: layout.DisplayOpFillRect, X: 10, Y: 4, Alpha: 1}
	base := cssmotion.Read(op)
	frames := []Keyframe{
		{Offset: 0, Transform: translateValue(0, 0)},
		{Offset: 1, Transform: translateValue(20, 10)},
	}
	p, err := New([]Target{{Op: op, From: base}}, frames, Timing{Duration: time.Second, Iterations: 1, TimingFunction: "linear"})
	if err != nil {
		t.Fatal(err)
	}
	start := time.Unix(30, 0)
	p.Start(start)
	p.Step(start.Add(500 * time.Millisecond))
	if !op.XformSet {
		t.Fatal("a transform keyframe did not write the operation matrix")
	}
	if op.X != base.X || op.Y != base.Y {
		t.Fatalf("matrix path rewrote the paint to %v,%v, want %v,%v", op.X, op.Y, base.X, base.Y)
	}
	m := op.Transform()
	if m.E != 10 || m.F != 5 {
		t.Fatalf("matrix translation = %v,%v, want 10,5", m.E, m.F)
	}
}

func TestComposeLeavesPositionForRotationOnly(t *testing.T) {
	op := &layout.DisplayOp{Kind: layout.DisplayOpFillRect, X: 10, Y: 4, Alpha: 1}
	base := cssmotion.Read(op)
	frames := []Keyframe{{Offset: 0, Transform: rotateValue(Rotate, 0, false)},
		{Offset: 1, Transform: rotateValue(Rotate, math.Pi/2, false)}}
	p, err := New([]Target{{Op: op, From: base}}, frames, Timing{Duration: time.Second, Iterations: 1})
	if err != nil {
		t.Fatal(err)
	}
	start := time.Unix(31, 0)
	p.Start(start)
	p.Step(start.Add(500 * time.Millisecond))
	if op.X != base.X || op.Y != base.Y {
		t.Fatalf("rotation moved the paint to %v,%v, want %v,%v", op.X, op.Y, base.X, base.Y)
	}
	if !op.XformSet {
		t.Fatal("a rotation keyframe did not write the operation matrix")
	}
	m := op.Transform()
	want := math.Pi / 4
	if math.Abs(m.A-math.Cos(want)) > matrixTolerance || math.Abs(m.B-math.Sin(want)) > matrixTolerance ||
		math.Abs(m.C+math.Sin(want)) > matrixTolerance || math.Abs(m.D-math.Cos(want)) > matrixTolerance ||
		m.E != 0 || m.F != 0 {
		t.Fatalf("rotation matrix = %+v, want 45 degrees about the origin", m)
	}
}

func TestNewRejectsMalformedTransform(t *testing.T) {
	bad := errors.New("bad transform")
	_, err := New([]Target{{Op: &layout.DisplayOp{}}}, []Keyframe{{Offset: 0, Err: bad}}, Timing{})
	if !errors.Is(err, bad) {
		t.Fatalf("New with a malformed keyframe = %v, want the parse error", err)
	}
}

func TestNewAcceptsMismatchedThreeDimensionalLists(t *testing.T) {
	op := &layout.DisplayOp{Kind: layout.DisplayOpFillRect, Alpha: 1}
	frames := []Keyframe{
		{Offset: 0, Transform: rotateValue(RotateX, 0, true)},
		{Offset: 1, Transform: rotateValue(RotateY, math.Pi/2, true)},
	}
	p, err := New([]Target{{Op: op, From: cssmotion.Read(op)}}, frames,
		Timing{Duration: time.Second, Iterations: 1, TimingFunction: "linear"})
	if err != nil {
		t.Fatalf("New with mismatched 3D lists = %v, want a bound player", err)
	}
	start := time.Unix(74, 0)
	p.Start(start)
	p.Step(start.Add(500 * time.Millisecond))

	if !p.RestampPending() {
		t.Fatal("a mismatched 3D sample did not reach the restamp path")
	}
	values := p.RestampValues()
	if len(values) != 1 {
		t.Fatalf("restamp values = %d, want one", len(values))
	}
	if !strings.HasPrefix(values[0].Transform, "matrix3d(") {
		t.Fatalf("restamp transform = %q, want matrix3d()", values[0].Transform)
	}
	if op.XformSet {
		t.Fatal("a 3D sample wrote a 2D operation matrix")
	}
}

func TestDecompose3DRejectsASingularMatrix(t *testing.T) {
	if _, err := Decompose3D(Matrix4{}); !errors.Is(err, ErrSingularMatrix) {
		t.Fatalf("zero matrix = %v, want ErrSingularMatrix", err)
	}
	scaleZ := Matrix4FromFunctions([]TransformFunction{{Name: ScaleZ, Args: []float64{0}}})
	if _, err := Decompose3D(scaleZ); !errors.Is(err, ErrSingularMatrix) {
		t.Fatalf("scaleZ(0) = %v, want ErrSingularMatrix", err)
	}
}
