package motion

import (
	"errors"
	"github.com/chinmay-sawant/blinkless/layout"
	"math"
	"testing"
	"time"
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

// matrix4ToArgs copies a matrix into the argument list of matrix3d().
func matrix4ToArgs(m Matrix4) []float64 {
	args := make([]float64, len(m))
	copy(args, m[:])
	return args
}

func identityMatrix3D() []float64 {
	return matrix4ToArgs(IdentityMatrix4())
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
		Functions:  []TransformFunction{{Name: Matrix3D, Args: matrix4ToArgs(rotation)}},
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
	base := Read(op)
	target := Target{Op: op, From: base, Parent: parent, World: world,
		OriginX: 10, OriginY: 20, HasElementTransform: true}

	p, err := newAnimationPlayer([]Target{target},
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
	from := Read(op)
	frames := []Keyframe{
		{Offset: 0, Transform: rotateValue(RotateY, 0, true)},
		{Offset: 1, Transform: rotateValue(RotateY, math.Pi/2, true)},
	}
	p, err := newAnimationPlayer([]Target{{Op: op, From: from}}, frames, Timing{Duration: time.Second, Iterations: 1})
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
	from := Read(op)
	target := Target{Op: op, From: from, Parent: layout.IdentityMatrix(),
		World: layout.IdentityMatrix(), HasTransform3D: true}
	frames := []Keyframe{
		{Offset: 0, Transform: translateValue(0, 0)},
		{Offset: 1, Transform: translateValue(40, 0)},
	}
	p, err := newAnimationPlayer([]Target{target}, frames, Timing{Duration: time.Second, Iterations: 1})
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
	from := Read(op)
	frames := []Keyframe{{Offset: 0, Move: true}, {Offset: 1, Move: true, X: 30}}
	p, err := newAnimationPlayer([]Target{{Op: op, From: from}}, frames,
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
	from := Read(op)
	frames := []Keyframe{
		{Offset: 0, Transform: rotateValue(Rotate, 0, false)},
		{Offset: 1, Transform: rotateValue(Rotate, math.Pi/2, false)},
	}
	p, err := newAnimationPlayer([]Target{{Op: op, From: from}}, frames, Timing{Duration: time.Second, Iterations: 1})
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

func testMatrices() []mat2 {
	return []mat2{
		identity2D(),
		translate2D(10, 20),
		scale2D(2, 3),
		scale2D(-1, 1),
		rotate2D(math.Pi / 3),
		skewX2D(math.Pi / 6),
		translate2D(4, 5).mul(rotate2D(math.Pi / 4)).mul(scale2D(2, 3)),
		translate2D(-7, 2).mul(rotate2D(-1.2)).mul(skewX2D(.3)).mul(scale2D(-2, .5)),
	}
}

func TestDecomposeRecomposeRoundTrip(t *testing.T) {
	for i, want := range testMatrices() {
		decomposed := decompose2D(want)
		got := recompose2D(decomposed)
		for _, field := range []struct {
			name string
			a, b float64
		}{
			{"a", want.a, got.a}, {"b", want.b, got.b},
			{"c", want.c, got.c}, {"d", want.d, got.d},
			{"e", want.e, got.e}, {"f", want.f, got.f},
		} {
			if math.Abs(field.a-field.b) > 1e-9 {
				t.Errorf("matrix %d %s = %v, want %v", i, field.name, field.b, field.a)
			}
		}
	}
}

func TestDecomposeTwoDimensionalFields(t *testing.T) {
	cases := []struct {
		name string
		m    mat2
		want Transform
	}{
		{"translate", translate2D(3, 4), Transform{Tx: 3, Ty: 4, Sx: 1, Sy: 1}},
		{"scale", scale2D(2, 3), Transform{Sx: 2, Sy: 3}},
		{"rotate", rotate2D(math.Pi / 2), Transform{Sx: 1, Sy: 1, Rotate: math.Pi / 2}},
		{"skewX", skewX2D(math.Pi / 6), Transform{Sx: 1, Sy: 1, SkewX: math.Pi / 6}},
	}
	for _, c := range cases {
		got := decompose2D(c.m)
		if math.Abs(got.Tx-c.want.Tx) > 1e-9 || math.Abs(got.Ty-c.want.Ty) > 1e-9 ||
			math.Abs(got.Sx-c.want.Sx) > 1e-9 || math.Abs(got.Sy-c.want.Sy) > 1e-9 ||
			math.Abs(got.Rotate-c.want.Rotate) > 1e-9 || math.Abs(got.SkewX-c.want.SkewX) > 1e-9 {
			t.Errorf("%s decomposed = %+v, want %+v", c.name, got, c.want)
		}
	}
}

func TestComposeTwoDimensionalListOrder(t *testing.T) {
	functions := []TransformFunction{
		{Name: Translate, Args: []float64{10, 20}},
		{Name: Scale, Args: []float64{2, 3}},
	}
	got := compose2D(functions)
	want := translate2D(10, 20).mul(scale2D(2, 3))
	if got != want {
		t.Fatalf("compose2D = %+v, want %+v", got, want)
	}
}

func TestComposeThreeDimensionalTranslation(t *testing.T) {
	functions := []TransformFunction{
		{Name: Translate3D, Args: []float64{10, 20, 30}},
	}
	m := compose3D(functions)
	if m[matrix3DTx] != 10 || m[matrix3DTy] != 20 || m[matrix3DTz] != 30 {
		t.Fatalf("translate3d column = %v,%v,%v, want 10,20,30",
			m[matrix3DTx], m[matrix3DTy], m[matrix3DTz])
	}
	rotated := compose3D([]TransformFunction{
		{Name: RotateZ, Args: []float64{math.Pi / 2}},
		{Name: TranslateX, Args: []float64{10}},
	})
	if math.Abs(rotated[matrix3DTx]) > 1e-9 || math.Abs(rotated[matrix3DTy]-10) > 1e-9 {
		t.Fatalf("rotate then translate = %v,%v, want 0,10", rotated[matrix3DTx], rotated[matrix3DTy])
	}
}

func TestPerspectiveMatrix(t *testing.T) {
	m := compose3D([]TransformFunction{{Name: Perspective, Args: []float64{100}}})
	if m[perspectiveW] != -1.0/100 {
		t.Fatalf("perspective slot = %v, want %v", m[perspectiveW], -1.0/100)
	}
}
