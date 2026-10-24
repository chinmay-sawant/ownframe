package animation

import (
	"errors"
	"math"
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

func TestInterpolateMismatchedThreeDimensionalListsReportsAnError(t *testing.T) {
	_, err := interpolateTransform(rotateValue(RotateX, 0, true), rotateValue(RotateY, 0, true), .5)
	if !errors.Is(err, ErrTransform3D) {
		t.Fatalf("mismatched 3D lists = %v, want ErrTransform3D", err)
	}
	_, err = interpolateTransform(rotateValue(RotateY, 0, true),
		TransformValue{Decomposed: IdentityTransform(), ThreeD: true,
			Functions: []TransformFunction{{Name: Perspective, Args: []float64{100}}}}, .5)
	if !errors.Is(err, ErrTransform3D) {
		t.Fatalf("perspective against none = %v, want ErrTransform3D", err)
	}
	matrix := TransformValue{Decomposed: IdentityTransform(), ThreeD: true,
		Functions: []TransformFunction{{Name: Matrix3D, Args: identityMatrix3D()}}}
	_, err = interpolateTransform(matrix, matrix, .5)
	if !errors.Is(err, ErrTransform3D) {
		t.Fatalf("matrix3d pair = %v, want ErrTransform3D", err)
	}
}

func identityMatrix3D() []float64 {
	return []float64{1, 0, 0, 0, 0, 1, 0, 0, 0, 0, 1, 0, 0, 0, 0, 1}
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

func TestNewRejectsMismatchedThreeDimensionalLists(t *testing.T) {
	frames := []Keyframe{
		{Offset: 0, Transform: rotateValue(RotateX, 0, true)},
		{Offset: 1, Transform: rotateValue(RotateY, math.Pi/2, true)},
	}
	_, err := New([]Target{{Op: &layout.DisplayOp{}}}, frames, Timing{})
	if !errors.Is(err, ErrTransform3D) {
		t.Fatalf("New with mismatched 3D lists = %v, want ErrTransform3D", err)
	}
}

func TestDecompose3DReportsTheMissingAlgorithm(t *testing.T) {
	from := []TransformFunction{{Name: RotateX, Args: []float64{0}}}
	to := []TransformFunction{{Name: RotateY, Args: []float64{0}}}
	if _, err := Decompose3D(from, to, .5); !errors.Is(err, ErrTransform3D) {
		t.Fatalf("Decompose3D = %v, want ErrTransform3D", err)
	}
	if err := Decompose3DError(from, to); !errors.Is(err, ErrTransform3D) {
		t.Fatalf("Decompose3DError = %v, want ErrTransform3D", err)
	}
}
