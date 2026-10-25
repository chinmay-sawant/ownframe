package animation

import (
	"math"
	"testing"
	"time"

	"github.com/chinmay-sawant/blinkless/layout"
	"github.com/chinmay-sawant/ownframe/internal/cssmotion"
)

func matrixNear(a, b layout.Matrix2D) bool {
	return math.Abs(a.A-b.A) <= matrixTolerance &&
		math.Abs(a.B-b.B) <= matrixTolerance &&
		math.Abs(a.C-b.C) <= matrixTolerance &&
		math.Abs(a.D-b.D) <= matrixTolerance &&
		math.Abs(a.E-b.E) <= matrixTolerance &&
		math.Abs(a.F-b.F) <= matrixTolerance
}

// rotateMatrix builds the CSS y-down rotation the engine stamps for rotate().
func rotateMatrix(deg float64) layout.Matrix2D {
	sin, cos := math.Sincos(deg * math.Pi / 180)
	return layout.Matrix2D{A: cos, B: sin, C: -sin, D: cos}
}

func TestInverseMatrix2DRoundTrips(t *testing.T) {
	m := layout.Matrix2D{A: 2, B: .5, C: -.25, D: 3, E: 7, F: -4}
	identity := layout.IdentityMatrix()
	if got := m.Mul(inverseMatrix2D(m)); !matrixNear(got, identity) {
		t.Fatalf("m * inverse(m) = %+v, want the identity", got)
	}
	if got := inverseMatrix2D(m).Mul(m); !matrixNear(got, identity) {
		t.Fatalf("inverse(m) * m = %+v, want the identity", got)
	}
}

func TestInverseMatrix2DSingularIsIdentity(t *testing.T) {
	singular := layout.Matrix2D{A: 1, B: 2, C: 2, D: 4, E: 3, F: 5}
	if got := inverseMatrix2D(singular); got != layout.IdentityMatrix() {
		t.Fatalf("singular inverse = %+v, want the identity", got)
	}
}

func TestMatrixFromTransformComposesTranslateRotateScale(t *testing.T) {
	// T(10,20) * R(90deg) * S(2,3):
	// R*S = [[0,-3],[2,0]], so the product is A=0, B=2, C=-3, D=0.
	got := matrixFromTransform(Transform{Tx: 10, Ty: 20, Sx: 2, Sy: 3, Rotate: math.Pi / 2, Set: true})
	want := layout.Matrix2D{A: 0, B: 2, C: -3, D: 0, E: 10, F: 20}
	if !matrixNear(got, want) {
		t.Fatalf("matrix = %+v, want %+v", got, want)
	}
}

func TestMatrixFromTransformAppliesSkewBeforeScale(t *testing.T) {
	// T(0,0) * R(0) * skewX(45deg) * S(2,1) has C = tan(45deg) * 1 = 1.
	got := matrixFromTransform(Transform{Sx: 2, Sy: 1, SkewX: math.Pi / 4, Set: true})
	if math.Abs(got.A-2) > matrixTolerance || math.Abs(got.C-1) > matrixTolerance ||
		math.Abs(got.D-1) > matrixTolerance {
		t.Fatalf("matrix = %+v, want A=2, C=1, D=1", got)
	}
}

func TestApplyTransformReplacesTheStaticWorld(t *testing.T) {
	op := &layout.DisplayOp{Kind: layout.DisplayOpFillRect, X: 10, Y: 20, W: 40, H: 20, Alpha: 1}
	parent := rotateMatrix(30)
	world := parent
	op.SetXform(world)
	base := cssmotion.Read(op)
	target := Target{Op: op, From: base, Parent: parent, World: world,
		OriginX: 10, OriginY: 20, HasElementTransform: true}

	p, err := New([]Target{target},
		[]Keyframe{
			{Offset: 0, Transform: rotateValue(Rotate, 0, false)},
			{Offset: 1, Transform: rotateValue(Rotate, math.Pi, false)},
		},
		Timing{Duration: time.Second, Iterations: 1, TimingFunction: "linear"})
	if err != nil {
		t.Fatal(err)
	}
	start := time.Unix(60, 0)
	p.Start(start)

	p.Step(start.Add(500 * time.Millisecond))

	want := parent.Mul(layout.BakeOrigin(rotateMatrix(90), 10, 20))
	if got := op.Transform(); !matrixNear(got, want) {
		t.Fatalf("mid matrix = %+v, want %+v", got, want)
	}
}

func TestApplyTransformSkipsAThreeDKeyframe(t *testing.T) {
	op := &layout.DisplayOp{Kind: layout.DisplayOpFillRect, X: 4, Y: 5, Alpha: 1}
	base := layout.Matrix2D{A: 1, D: 1, E: 6}
	op.SetXform(base)
	from := cssmotion.Read(op)
	frames := []Keyframe{
		{Offset: 0, Transform: rotateValue(RotateY, 0, true)},
		{Offset: 1, Transform: rotateValue(RotateY, math.Pi/2, true)},
	}
	p, err := New([]Target{{Op: op, From: from}}, frames, Timing{Duration: time.Second, Iterations: 1})
	if err != nil {
		t.Fatal(err)
	}
	start := time.Unix(61, 0)
	p.Start(start)
	p.Step(start.Add(500 * time.Millisecond))

	if got := op.Transform(); got != base {
		t.Fatalf("3D keyframe wrote matrix %+v, want the base %+v left alone", got, base)
	}
	if !p.RestampPending() {
		t.Fatal("a 3D keyframe did not ask for the engine restamp pass")
	}
}

func TestApplyTransformSkipsAThreeDElement(t *testing.T) {
	op := &layout.DisplayOp{Kind: layout.DisplayOpFillRect, X: 4, Y: 5, Alpha: 1}
	from := cssmotion.Read(op)
	target := Target{Op: op, From: from, Parent: layout.IdentityMatrix(),
		World: layout.IdentityMatrix(), HasTransform3D: true}
	frames := []Keyframe{
		{Offset: 0, Transform: translateValue(0, 0)},
		{Offset: 1, Transform: translateValue(40, 0)},
	}
	p, err := New([]Target{target}, frames, Timing{Duration: time.Second, Iterations: 1})
	if err != nil {
		t.Fatal(err)
	}
	start := time.Unix(62, 0)
	p.Start(start)
	p.Step(start.Add(500 * time.Millisecond))

	if op.XformSet {
		t.Fatal("a 3D element received a 2D matrix")
	}
	if !p.RestampPending() {
		t.Fatal("a 3D element did not ask for the engine restamp pass")
	}
}

func TestApplyTransformKeepsLegacyMoveOnTheScalarPath(t *testing.T) {
	op := &layout.DisplayOp{Kind: layout.DisplayOpFillRect, X: 10, Y: 4, Alpha: 1}
	from := cssmotion.Read(op)
	frames := []Keyframe{{Offset: 0, Move: true}, {Offset: 1, Move: true, X: 30}}
	p, err := New([]Target{{Op: op, From: from}}, frames,
		Timing{Duration: time.Second, Iterations: 1, TimingFunction: "linear"})
	if err != nil {
		t.Fatal(err)
	}
	start := time.Unix(63, 0)
	p.Start(start)
	p.Step(start.Add(500 * time.Millisecond))

	if op.XformSet {
		t.Fatal("a legacy X/Y keyframe wrote a matrix")
	}
	if op.X != from.X+15 {
		t.Fatalf("legacy X = %v, want the base %v plus 15", op.X, from.X)
	}
}

func TestRestampPendingIsFalseForTwoDimensionalSamples(t *testing.T) {
	op := &layout.DisplayOp{Kind: layout.DisplayOpFillRect, X: 4, Y: 5, Alpha: 1}
	from := cssmotion.Read(op)
	frames := []Keyframe{
		{Offset: 0, Transform: rotateValue(Rotate, 0, false)},
		{Offset: 1, Transform: rotateValue(Rotate, math.Pi/2, false)},
	}
	p, err := New([]Target{{Op: op, From: from}}, frames, Timing{Duration: time.Second, Iterations: 1})
	if err != nil {
		t.Fatal(err)
	}
	start := time.Unix(64, 0)
	p.Start(start)
	p.Step(start.Add(500 * time.Millisecond))

	if p.RestampPending() {
		t.Fatal("a 2D transform asked for the restamp pass")
	}
}
