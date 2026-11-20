package motion

import (
	"errors"
	"github.com/chinmay-sawant/blinkless/layout"
	"math"
	"strconv"
	"strings"
	"testing"
	"time"
)

// TestAdditiveMissingKeyframeAddsNothing checks that a keyframe missing a
// property contributes nothing under an additive composition. The
// underlying paint must survive at 0% and 100% instead of being added to
// itself and saturating.
func TestAdditiveMissingKeyframeAddsNothing(t *testing.T) {
	op := &layout.DisplayOp{Kind: layout.DisplayOpFillRect, X: 8, Y: 6,
		R: .2, G: .4, B: .6, Alpha: 1}
	op.SetPaintOpacity(.5)
	base := Read(op)
	p, err := newAnimationPlayer([]Target{{Op: op, From: base}}, []Keyframe{
		{Offset: .5, Move: true, X: 40, Opacity: .25, OpacitySet: true,
			R: .3, G: .3, B: .3, ColorSet: true},
	}, Timing{Duration: time.Second, Iterations: 1, Fill: FillBoth,
		TimingFunction: "linear", Composition: Add})
	if err != nil {
		t.Fatal(err)
	}
	start := time.Unix(80, 0)
	p.Start(start)
	p.Step(start)
	checkAdditivePaint(t, op, base, "start")
	p.Step(start.Add(500 * time.Millisecond))
	mid := base
	mid.R, mid.G, mid.B = base.R+.3, base.G+.3, base.B+.3
	mid.Opacity = base.Opacity + .25
	checkAdditivePaint(t, op, mid, "mid")
	if op.X != base.X+40 {
		t.Fatalf("mid X = %v, want %v", op.X, base.X+40)
	}
	p.Step(start.Add(3 * time.Second))
	checkAdditivePaint(t, op, base, "end")
	if op.X != base.X || op.Y != base.Y {
		t.Fatalf("end position = %v,%v, want %v,%v", op.X, op.Y, base.X, base.Y)
	}
}

// checkAdditivePaint reports whether the op holds the wanted paint.
func checkAdditivePaint(t *testing.T, op *layout.DisplayOp, want Paint, when string) {
	t.Helper()
	if math.Abs(op.R-want.R) > .001 || math.Abs(op.G-want.G) > .001 ||
		math.Abs(op.B-want.B) > .001 {
		t.Fatalf("%s color = %v,%v,%v, want %v,%v,%v",
			when, op.R, op.G, op.B, want.R, want.G, want.B)
	}
	if math.Abs(op.Opacity()-want.Opacity) > .001 {
		t.Fatalf("%s opacity = %v, want %v", when, op.Opacity(), want.Opacity)
	}
}

func TestAccumulateComposesCurrentValueOnLaterIteration(t *testing.T) {
	op := &layout.DisplayOp{Kind: layout.DisplayOpFillRect, X: 12, Alpha: 1}
	base := Read(op)
	p, err := newAnimationPlayer([]Target{{Op: op, From: base}}, []Keyframe{
		{Offset: 0, X: 4, Move: true}, {Offset: 1, X: 10, Move: true},
	}, Timing{Duration: time.Second, Iterations: 2, Fill: FillBoth, Composition: Accumulate})
	if err != nil {
		t.Fatal(err)
	}
	start := time.Unix(60, 0)
	p.Start(start)
	p.Step(start.Add(1500 * time.Millisecond))
	if want := base.X + 7; math.Abs(op.X-want) > .001 {
		t.Fatalf("accumulated X = %v, want underlying plus current value %v", op.X, want)
	}
}

// constantFrames holds one transform for the whole active interval, so a
// composition test can isolate the effect stack from keyframe interpolation.
func constantFrames(value TransformValue) []Keyframe {
	return []Keyframe{{Offset: 0, Transform: value}, {Offset: 1, Transform: value}}
}

// compositionTiming is a one-second linear effect with the given mode.
func compositionTiming(mode Composition) Timing {
	return Timing{Duration: time.Second, Iterations: 1, TimingFunction: "linear", Composition: mode}
}

// stepComposition binds one target to several effects and samples the midpoint
// of the first iteration.
func stepComposition(t *testing.T, target Target, effects []Effect) (*Player, *layout.DisplayOp) {
	t.Helper()
	p, err := NewComposite([]Target{target}, effects)
	if err != nil {
		t.Fatal(err)
	}
	start := time.Unix(90, 0)
	p.Start(start)
	p.Step(start.Add(500 * time.Millisecond))

	return p, target.Op
}

// matrix3DFromCSS reads a canonical matrix3d() list the restamp path produced.
// All sixteen arguments are unitless; the translation components are already
// in display points.
func matrix3DFromCSS(t *testing.T, css string) Matrix4 {
	t.Helper()
	body, ok := strings.CutPrefix(css, "matrix3d(")
	if !ok || !strings.HasSuffix(body, ")") {
		t.Fatalf("restamp transform = %q, want matrix3d()", css)
	}
	parts := strings.Split(strings.TrimSuffix(body, ")"), ", ")
	var m Matrix4
	if len(parts) != len(m) {
		t.Fatalf("matrix3d has %d arguments, want %d", len(parts), len(m))
	}
	for i, part := range parts {
		value, err := strconv.ParseFloat(part, 64)
		if err != nil {
			t.Fatalf("matrix3d argument %d = %q: %v", i, part, err)
		}
		m[i] = value
	}

	return m
}

// TestAddConcatenatesTwoTranslateEffects checks the spec case where two
// additive translate effects on one element compose instead of the later one
// replacing the earlier one.
func TestAddConcatenatesTwoTranslateEffects(t *testing.T) {
	op := &layout.DisplayOp{Kind: layout.DisplayOpFillRect, X: 4, Alpha: 1}
	base := Read(op)
	p, op := stepComposition(t, Target{Op: op, From: base}, []Effect{
		{Frames: constantFrames(translateValue(10, 0)), Timing: compositionTiming(Add)},
		{Frames: constantFrames(translateValue(20, 0)), Timing: compositionTiming(Add)},
	})
	if p.RestampPending() {
		t.Fatal("a two-dimensional add asked for the restamp pass")
	}
	if got := op.Transform().E; math.Abs(got-30) > matrixTolerance {
		t.Fatalf("translated x = %v, want 30 from concatenating 10 and 20", got)
	}
}

// TestAddConcatenatesTranslateAndRotate checks that an additive 3D sample
// composes after the 2D transform already sampled for the element.
func TestAddConcatenatesTranslateAndRotate(t *testing.T) {
	op := &layout.DisplayOp{Kind: layout.DisplayOpFillRect, X: 4, Alpha: 1}
	base := Read(op)
	p, _ := stepComposition(t, Target{Op: op, From: base}, []Effect{
		{Frames: constantFrames(translateValue(10, 0)), Timing: compositionTiming(Add)},
		{Frames: constantFrames(rotateValue(RotateY, math.Pi/2, true)), Timing: compositionTiming(Add)},
	})
	values := p.RestampValues()
	if len(values) != 1 {
		t.Fatalf("restamp values = %d, want one composed 3D transform", len(values))
	}
	got := matrix3DFromCSS(t, values[0].Transform)
	want := translation4(10, 0, 0).Mul(rotateY4(math.Pi / 2))
	requireMatrix4Near(t, got, want, matrixTolerance)
}

// TestAccumulateCombinesComponentsOverTwoIterations checks the accumulation
// procedure against add over a second iteration. Both modes add the underlying
// value once per sample, but add concatenates (multiplies) the lists while
// accumulate adds the decomposed components with the one-based scale rule.
// animation-composition is the composite operation, not iterationComposite, so
// the second iteration accumulates the same underlying value instead of
// compounding.
func TestAccumulateCombinesComponentsOverTwoIterations(t *testing.T) {
	for _, test := range []struct {
		name string
		mode Composition
		want float64
	}{
		{name: "add", mode: Add, want: 2},
		{name: "accumulate", mode: Accumulate, want: 1.5},
	} {
		t.Run(test.name, func(t *testing.T) {
			op := &layout.DisplayOp{Kind: layout.DisplayOpFillRect, Alpha: 1}
			base := Read(op)
			// The element's own transform is scale(2), so the composed
			// animation matrix is visible as a factor of the world matrix.
			target := Target{Op: op, From: base, World: layout.Matrix2D{A: 2, D: 2}}
			frames := []Keyframe{
				{Offset: 0, Transform: scaleValue(1)},
				{Offset: 1, Transform: scaleValue(3)},
			}
			timing := Timing{Duration: time.Second, Iterations: 2, TimingFunction: "linear",
				Composition: test.mode}
			p, err := newAnimationPlayer([]Target{target}, frames, timing)
			if err != nil {
				t.Fatal(err)
			}
			start := time.Unix(91, 0)
			p.Start(start)

			for _, at := range []time.Duration{500 * time.Millisecond, 1500 * time.Millisecond} {
				p.Step(start.Add(at))
				if got := op.Transform().A; math.Abs(got-test.want) > matrixTolerance {
					t.Fatalf("scale at %v = %v, want %v", at, got, test.want)
				}
			}
		})
	}
}

// TestAccumulateComposesThreeDimensionalRotations checks the 4x4 accumulation
// path: translations add, one-based scales add, and rotations multiply their
// quaternions.
func TestAccumulateComposesThreeDimensionalRotations(t *testing.T) {
	underlying := TransformValue{Decomposed: Transform{Rotate: math.Pi / 4, Sx: 1, Sy: 1, Set: true}}
	composed := composeTransform3D(underlying, rotateValue(RotateY, math.Pi/2, true), Accumulate)
	got := Matrix4FromFunctions(composed.Functions)
	want := rotateZ4(math.Pi / 4).Mul(rotateY4(math.Pi / 2))
	requireMatrix4Near(t, got, want, matrixTolerance)
}

// TestTransformSegmentUsesKeyframeEasing checks that the keyframe that starts
// a transform interval supplies its easing, not the effect-level function.
func TestTransformSegmentUsesKeyframeEasing(t *testing.T) {
	op := &layout.DisplayOp{Kind: layout.DisplayOpFillRect, Alpha: 1}
	base := Read(op)
	frames := []Keyframe{
		{Offset: 0, Transform: translateValue(0, 0), TimingFunction: "ease-in"},
		{Offset: 1, Transform: translateValue(100, 0), TimingFunction: "linear"},
	}
	p, err := newAnimationPlayer([]Target{{Op: op, From: base}}, frames,
		Timing{Duration: time.Second, Iterations: 1, TimingFunction: "linear"})
	if err != nil {
		t.Fatal(err)
	}
	start := time.Unix(92, 0)
	p.Start(start)
	p.Step(start.Add(500 * time.Millisecond))

	eased := Ease("ease-in", 0.5)
	if math.Abs(eased-0.5) < matrixTolerance {
		t.Fatal("ease-in has a linear midpoint; the test cannot see the override")
	}
	if got, want := op.Transform().E, 100*eased; math.Abs(got-want) > matrixTolerance {
		t.Fatalf("translate midpoint = %v, want %v from the keyframe easing", got, want)
	}
}

// TestTransformSegmentEasingFollowsTheStartingKeyframe checks a three-stop
// list where each segment has its own function.
func TestTransformSegmentEasingFollowsTheStartingKeyframe(t *testing.T) {
	op := &layout.DisplayOp{Kind: layout.DisplayOpFillRect, Alpha: 1}
	base := Read(op)
	frames := []Keyframe{
		{Offset: 0, Transform: translateValue(0, 0), TimingFunction: "linear"},
		{Offset: 0.5, Transform: translateValue(40, 0), TimingFunction: "ease-in"},
		{Offset: 1, Transform: translateValue(80, 0), TimingFunction: "linear"},
	}
	p, err := newAnimationPlayer([]Target{{Op: op, From: base}}, frames,
		Timing{Duration: time.Second, Iterations: 1, TimingFunction: "linear"})
	if err != nil {
		t.Fatal(err)
	}
	start := time.Unix(93, 0)
	p.Start(start)
	p.Step(start.Add(750 * time.Millisecond))

	want := 40 + 40*Ease("ease-in", 0.5)
	if got := op.Transform().E; math.Abs(got-want) > matrixTolerance {
		t.Fatalf("second segment = %v, want %v from ease-in", got, want)
	}
}

// TestPerspectiveSegmentUsesKeyframeEasing checks the same rule on the
// perspective property.
func TestPerspectiveSegmentUsesKeyframeEasing(t *testing.T) {
	op := &layout.DisplayOp{Kind: layout.DisplayOpFillRect, Alpha: 1}
	base := Read(op)
	frames := []Keyframe{
		{Offset: 0, Perspective: cameraValue(500, true), TimingFunction: "ease-in"},
		{Offset: 1, Perspective: cameraValue(1000, true), TimingFunction: "linear"},
	}
	p, err := newAnimationPlayer([]Target{{Op: op, From: base}}, frames, perspectiveTiming())
	if err != nil {
		t.Fatal(err)
	}
	start := time.Unix(94, 0)
	p.Start(start)
	p.Step(start.Add(500 * time.Millisecond))

	values := p.RestampValues()
	if len(values) != 1 || !values[0].HasPerspective {
		t.Fatalf("restamp values = %+v, want one perspective sample", values)
	}
	want := 500 + 500*Ease("ease-in", 0.5)
	if got := values[0].Perspective; math.Abs(got-want) > perspectiveTolerance {
		t.Fatalf("perspective midpoint = %v, want %v from the keyframe easing", got, want)
	}
}

// TestMixedCompositionKeepsDeclarationOrder checks that replace, add, and
// accumulate fold in list order. The rotate between the translates makes the
// order observable in the final translation.
func TestMixedCompositionKeepsDeclarationOrder(t *testing.T) {
	op := &layout.DisplayOp{Kind: layout.DisplayOpFillRect, Alpha: 1}
	base := Read(op)
	effects := []Effect{
		{Frames: constantFrames(translateValue(10, 0)), Timing: compositionTiming(Replace)},
		{Frames: constantFrames(rotateValue(Rotate, math.Pi/2, false)), Timing: compositionTiming(Add)},
		{Frames: constantFrames(translateValue(5, 0)), Timing: compositionTiming(Accumulate)},
	}
	_, op = stepComposition(t, Target{Op: op, From: base}, effects)

	// T(10) * R(90deg) accumulated with T(5) is T(15) * R(90deg).
	want := layout.Matrix2D{A: 0, B: 1, C: -1, D: 0, E: 15}
	if got := op.Transform(); !matrixNear(got, want) {
		t.Fatalf("mixed matrix = %+v, want %+v", got, want)
	}
}

// TestReplaceClobbersEarlierTransformEffects checks that a later replace does
// not keep an earlier list, whatever the earlier mode was.
func TestReplaceClobbersEarlierTransformEffects(t *testing.T) {
	op := &layout.DisplayOp{Kind: layout.DisplayOpFillRect, Alpha: 1}
	base := Read(op)
	effects := []Effect{
		{Frames: constantFrames(translateValue(10, 0)), Timing: compositionTiming(Add)},
		{Frames: constantFrames(translateValue(5, 0)), Timing: compositionTiming(Replace)},
	}
	_, op = stepComposition(t, Target{Op: op, From: base}, effects)

	if got := op.Transform().E; math.Abs(got-5) > matrixTolerance {
		t.Fatalf("replace x = %v, want 5", got)
	}
}

// TestSingleEffectIsUnchanged checks that one effect with no underlying
// transform installs exactly its own transform in every composition mode, and
// that a single 3D sample keeps the authored list instead of a matrix3d() round
// trip.
func TestSingleEffectIsUnchanged(t *testing.T) {
	for _, mode := range []Composition{Replace, Add, Accumulate} {
		t.Run(compositionName(mode), func(t *testing.T) {
			op := &layout.DisplayOp{Kind: layout.DisplayOpFillRect, Alpha: 1}
			base := Read(op)
			value := translateValue(12, -4)
			p, op := stepComposition(t, Target{Op: op, From: base}, []Effect{
				{Frames: constantFrames(value), Timing: compositionTiming(mode)},
			})
			if p.RestampPending() {
				t.Fatal("a single 2D effect asked for the restamp pass")
			}
			if got, want := op.Transform(), matrixFromTransform(value.Decomposed); !matrixNear(got, want) {
				t.Fatalf("single effect matrix = %+v, want %+v", got, want)
			}

			op = &layout.DisplayOp{Kind: layout.DisplayOpFillRect, Alpha: 1}
			base = Read(op)
			p, _ = stepComposition(t, Target{Op: op, From: base}, []Effect{
				{Frames: []Keyframe{
					{Offset: 0, Transform: rotateValue(RotateY, 0, true)},
					{Offset: 1, Transform: rotateValue(RotateY, math.Pi/2, true)},
				}, Timing: compositionTiming(mode)},
			})
			values := p.RestampValues()
			if len(values) != 1 || values[0].Transform != "rotateY(45deg)" {
				t.Fatalf("single 3D effect serialized %+v, want the authored rotateY list", values)
			}
		})
	}
}

// compositionName labels a mode in a subtest name.
func compositionName(mode Composition) string {
	switch mode {
	case Add:
		return "add"
	case Accumulate:
		return "accumulate"
	case Replace:
		return "replace"
	}

	return "unknown"
}

// scaleValue builds a keyframe scale(scale) in both representations.
func scaleValue(scale float64) TransformValue {
	return TransformValue{
		Decomposed: Transform{Sx: scale, Sy: scale, Set: true},
		Functions:  []TransformFunction{{Name: Scale, Args: []float64{scale, scale}}},
	}
}

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
		Functions:  []TransformFunction{{Name: Matrix3D, Args: matrix4ToArgs(rotation)}},
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
	base := Read(op)
	frames := []Keyframe{
		{Offset: 0, Transform: translateValue(0, 0)},
		{Offset: 1, Transform: translateValue(20, 10)},
	}
	p, err := newAnimationPlayer([]Target{{Op: op, From: base}}, frames, Timing{Duration: time.Second, Iterations: 1, TimingFunction: "linear"})
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
	base := Read(op)
	frames := []Keyframe{{Offset: 0, Transform: rotateValue(Rotate, 0, false)},
		{Offset: 1, Transform: rotateValue(Rotate, math.Pi/2, false)}}
	p, err := newAnimationPlayer([]Target{{Op: op, From: base}}, frames, Timing{Duration: time.Second, Iterations: 1})
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
	_, err := newAnimationPlayer([]Target{{Op: &layout.DisplayOp{}}}, []Keyframe{{Offset: 0, Err: bad}}, Timing{})
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
	p, err := newAnimationPlayer([]Target{{Op: op, From: Read(op)}}, frames,
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

const perspectiveTolerance = 1e-9

// cameraValue builds a perspective declaration with the test element's default
// 50%/50% origin already resolved to 50 and 25 points.
func cameraValue(distance float64, hasDistance bool) PerspectiveValue {
	return PerspectiveValue{
		Distance:    distance,
		OriginX:     50,
		OriginY:     25,
		HasDistance: hasDistance,
	}
}

func cameraOriginValue(distance, x, y float64) PerspectiveValue {
	return PerspectiveValue{
		Distance:    distance,
		OriginX:     x,
		OriginY:     y,
		HasDistance: true,
		HasOrigin:   true,
	}
}

func perspectiveTiming() Timing {
	return Timing{Duration: time.Second, Iterations: 1, TimingFunction: "linear"}
}

func TestPerspectiveDistanceInterpolatesLinearly(t *testing.T) {
	op := &layout.DisplayOp{Kind: layout.DisplayOpFillRect, Alpha: 1}
	p, err := newAnimationPlayer([]Target{{Op: op, From: Read(op)}}, []Keyframe{
		{Offset: 0, Perspective: cameraValue(500, true)},
		{Offset: 1, Perspective: cameraValue(1000, true)},
	}, perspectiveTiming())
	if err != nil {
		t.Fatal(err)
	}
	start := time.Unix(80, 0)
	p.Start(start)
	p.Step(start.Add(500 * time.Millisecond))

	values := p.RestampValues()
	if len(values) != 1 {
		t.Fatalf("restamp values = %d, want one perspective sample", len(values))
	}
	if !values[0].HasPerspective || math.Abs(values[0].Perspective-750) > perspectiveTolerance {
		t.Fatalf("perspective = %+v, want 750 points at the midpoint", values[0])
	}
	if values[0].Transform != "" {
		t.Fatalf("perspective-only sample carried transform %q", values[0].Transform)
	}
	if values[0].HasPerspectiveOrigin {
		t.Fatalf("sample reported an origin the keyframes did not declare: %+v", values[0])
	}
	if !p.RestampPending() {
		t.Fatal("a perspective sample did not mark the restamp")
	}
	if op.XformSet {
		t.Fatal("a perspective sample wrote a 2D operation matrix")
	}
}

func TestPerspectiveNoneIsDiscreteAtTheMidpoint(t *testing.T) {
	effect := Effect{
		Frames: []Keyframe{
			{Offset: 0, Perspective: cameraValue(0, true)},
			{Offset: 1, Perspective: cameraValue(500, true)},
		},
		Timing: perspectiveTiming(),
	}

	before, ok := effect.samplePerspective(0.49)
	if !ok {
		t.Fatal("the none-to-length effect reported no perspective")
	}
	if before.Distance != 0 || !before.HasDistance {
		t.Fatalf("at 49%% perspective = %+v, want the none endpoint", before)
	}
	if !before.Discrete {
		t.Fatal("a none endpoint did not report a discrete sample")
	}

	after, ok := effect.samplePerspective(0.51)
	if !ok {
		t.Fatal("the none-to-length effect reported no perspective")
	}
	if after.Distance != 500 || !after.HasDistance {
		t.Fatalf("at 51%% perspective = %+v, want the length endpoint", after)
	}
	if !after.Discrete {
		t.Fatal("a none endpoint did not report a discrete sample")
	}

	mid, _ := effect.samplePerspective(0.5)
	if mid.Distance != 500 {
		t.Fatalf("at 50%% perspective = %v, want the length endpoint", mid.Distance)
	}
}

func TestPerspectiveOriginInterpolatesComponentWise(t *testing.T) {
	effect := Effect{
		Frames: []Keyframe{
			{Offset: 0, Perspective: cameraOriginValue(500, 0, 0)},
			{Offset: 1, Perspective: cameraOriginValue(500, 100, 50)},
		},
		Timing: perspectiveTiming(),
	}

	value, ok := effect.samplePerspective(0.25)
	if !ok {
		t.Fatal("the origin effect reported no perspective")
	}
	if !value.HasOrigin || math.Abs(value.OriginX-25) > perspectiveTolerance ||
		math.Abs(value.OriginY-12.5) > perspectiveTolerance {
		t.Fatalf("origin = %+v, want 25,12.5 at 25%%", value)
	}
	if math.Abs(value.Distance-500) > perspectiveTolerance {
		t.Fatalf("distance = %v, want the constant 500", value.Distance)
	}
}

func TestPerspectiveOriginUsesTheUndeclaredDefault(t *testing.T) {
	effect := Effect{
		Frames: []Keyframe{
			{Offset: 0, Perspective: cameraValue(500, true)},
			{Offset: 1, Perspective: cameraOriginValue(500, 0, 0)},
		},
		Timing: perspectiveTiming(),
	}

	value, ok := effect.samplePerspective(0.5)
	if !ok {
		t.Fatal("the origin effect reported no perspective")
	}
	// The undeclared 50%/50% endpoint carries the frame's resolved default.
	if !value.HasOrigin || math.Abs(value.OriginX-25) > perspectiveTolerance ||
		math.Abs(value.OriginY-12.5) > perspectiveTolerance {
		t.Fatalf("origin = %+v, want the midpoint between 50,25 and 0,0", value)
	}
}

func TestPerspectiveAndThreeDimensionalTransformShareOneSample(t *testing.T) {
	op := &layout.DisplayOp{Kind: layout.DisplayOpFillRect, Alpha: 1}
	p, err := newAnimationPlayer([]Target{{Op: op, From: Read(op)}}, []Keyframe{
		{Offset: 0, Transform: rotateValue(RotateY, 0, true), Perspective: cameraValue(500, true)},
		{Offset: 1, Transform: rotateValue(RotateY, math.Pi/2, true), Perspective: cameraValue(1000, true)},
	}, perspectiveTiming())
	if err != nil {
		t.Fatal(err)
	}
	start := time.Unix(81, 0)
	p.Start(start)
	p.Step(start.Add(500 * time.Millisecond))

	values := p.RestampValues()
	if len(values) != 1 {
		t.Fatalf("restamp values = %d, want one for the element", len(values))
	}
	if values[0].Transform != "rotateY(45deg)" {
		t.Fatalf("transform = %q, want the sampled rotateY", values[0].Transform)
	}
	if !values[0].HasPerspective || math.Abs(values[0].Perspective-750) > perspectiveTolerance {
		t.Fatalf("perspective = %+v, want the sampled distance", values[0])
	}
}

func TestWantsRestampSeesAPerspectiveKeyframe(t *testing.T) {
	threeD := []Keyframe{
		{Offset: 0, Transform: rotateValue(RotateY, 0, true)},
		{Offset: 1, Transform: rotateValue(RotateY, math.Pi, true)},
	}
	perspective := []Keyframe{
		{Offset: 0, Perspective: cameraValue(500, true)},
		{Offset: 1, Perspective: cameraValue(1000, true)},
	}
	originOnly := []Keyframe{
		{Offset: 0, Perspective: cameraOriginValue(0, 0, 0)},
		{Offset: 1, Perspective: cameraOriginValue(0, 100, 50)},
	}
	twoD := []Keyframe{
		{Offset: 0, Transform: translateValue(0, 0)},
		{Offset: 1, Transform: translateValue(10, 0)},
	}

	for _, test := range []struct {
		name    string
		frames  []Keyframe
		expects bool
	}{
		{name: "3D keyframe", frames: threeD, expects: true},
		{name: "perspective distance", frames: perspective, expects: true},
		{name: "perspective origin", frames: originOnly, expects: true},
		{name: "2D keyframe", frames: twoD, expects: false},
	} {
		t.Run(test.name, func(t *testing.T) {
			p, err := newAnimationPlayer([]Target{{Op: &layout.DisplayOp{}}}, test.frames, perspectiveTiming())
			if err != nil {
				t.Fatal(err)
			}
			if got := p.WantsRestamp(); got != test.expects {
				t.Fatalf("WantsRestamp = %v, want %v", got, test.expects)
			}
		})
	}
}

func TestPerspectiveOriginOnlySampleLeavesTheDistanceAlone(t *testing.T) {
	value := cameraOriginValue(0, 0, 0)
	value.HasDistance = false
	op := &layout.DisplayOp{Kind: layout.DisplayOpFillRect, Alpha: 1}
	p, err := newAnimationPlayer([]Target{{Op: op, From: Read(op)}}, []Keyframe{
		{Offset: 0, Perspective: value},
		{Offset: 1, Perspective: PerspectiveValue{OriginX: 100, OriginY: 50, HasOrigin: true}},
	}, perspectiveTiming())
	if err != nil {
		t.Fatal(err)
	}
	start := time.Unix(82, 0)
	p.Start(start)
	p.Step(start.Add(250 * time.Millisecond))

	values := p.RestampValues()
	if len(values) != 1 {
		t.Fatalf("restamp values = %d, want one", len(values))
	}
	if values[0].HasPerspective {
		t.Fatalf("origin-only sample reports HasPerspective: %+v", values[0])
	}
	if !values[0].HasPerspectiveOrigin {
		t.Fatalf("origin-only sample lost the origin: %+v", values[0])
	}
	if math.Abs(values[0].PerspectiveOriginX-25) > perspectiveTolerance ||
		math.Abs(values[0].PerspectiveOriginY-12.5) > perspectiveTolerance {
		t.Fatalf("origin = %v,%v, want 25,12.5", values[0].PerspectiveOriginX, values[0].PerspectiveOriginY)
	}
}

func TestTwoDimensionalSampleReportsNoPerspective(t *testing.T) {
	op := &layout.DisplayOp{Kind: layout.DisplayOpFillRect, Alpha: 1}
	p, err := newAnimationPlayer([]Target{{Op: op, From: Read(op)}}, []Keyframe{
		{Offset: 0, Transform: rotateValue(RotateY, 0, true)},
		{Offset: 1, Transform: rotateValue(RotateY, math.Pi/2, true)},
	}, perspectiveTiming())
	if err != nil {
		t.Fatal(err)
	}
	start := time.Unix(83, 0)
	p.Start(start)
	p.Step(start.Add(500 * time.Millisecond))

	values := p.RestampValues()
	if len(values) != 1 {
		t.Fatalf("restamp values = %d, want the 3D transform", len(values))
	}
	if values[0].HasPerspective || values[0].HasPerspectiveOrigin {
		t.Fatalf("a 3D-only sample reported perspective fields: %+v", values[0])
	}
}

// cameraBase is the static camera a keyframe list without a 0% or 100%
// perspective stop holds.
func cameraBase(distance float64, hasDistance bool) PerspectiveBase {
	return PerspectiveBase{
		Distance:    distance,
		OriginX:     7,
		OriginY:     9,
		HasDistance: hasDistance,
		Set:         true,
	}
}

// TestPerspectiveBasePadsAPartialList checks that the sampler brackets a
// partial keyframe list with the element's static distance, not none.
func TestPerspectiveBasePadsAPartialList(t *testing.T) {
	base := cameraBase(800, true)
	effect := Effect{
		Frames: []Keyframe{{
			Offset: 0.5,
			Perspective: PerspectiveValue{
				Distance: 500, OriginX: 7, OriginY: 9, HasDistance: true,
			},
			PerspectiveBase: base,
		}},
		Timing: perspectiveTiming(),
	}

	start, ok := effect.samplePerspective(0)
	if !ok {
		t.Fatal("the partial effect reported no perspective")
	}
	if !start.HasDistance || math.Abs(start.Distance-800) > perspectiveTolerance {
		t.Fatalf("0%% perspective = %+v, want the static 800", start)
	}
	if math.Abs(start.OriginX-7) > perspectiveTolerance ||
		math.Abs(start.OriginY-9) > perspectiveTolerance {
		t.Fatalf("0%% origin = %v,%v, want the static 7,9", start.OriginX, start.OriginY)
	}

	middle, _ := effect.samplePerspective(0.5)
	if math.Abs(middle.Distance-500) > perspectiveTolerance {
		t.Fatalf("50%% perspective = %v, want the declared 500", middle.Distance)
	}
}

// TestPerspectiveBaseNoneKeepsNone checks that a static none stays a none
// endpoint instead of becoming the box's default camera.
func TestPerspectiveBaseNoneKeepsNone(t *testing.T) {
	base := cameraBase(0, false)
	effect := Effect{
		Frames: []Keyframe{{
			Offset: 0.5,
			Perspective: PerspectiveValue{
				Distance: 500, OriginX: 7, OriginY: 9, HasDistance: true,
			},
			PerspectiveBase: base,
		}},
		Timing: perspectiveTiming(),
	}

	value, ok := effect.samplePerspective(0)
	if !ok {
		t.Fatal("the partial effect reported no perspective")
	}
	if !value.HasDistance || math.Abs(value.Distance) > perspectiveTolerance {
		t.Fatalf("0%% perspective = %+v, want a none endpoint", value)
	}

	early, _ := effect.samplePerspective(0.1)
	if math.Abs(early.Distance) > perspectiveTolerance {
		t.Fatalf("10%% perspective = %v, want the none endpoint before the midpoint", early.Distance)
	}
}

// TestExplicitPerspectiveEndpointBeatsTheBase checks that a declared endpoint
// is not replaced by the static base.
func TestExplicitPerspectiveEndpointBeatsTheBase(t *testing.T) {
	base := cameraBase(800, true)
	effect := Effect{
		Frames: []Keyframe{
			{Offset: 0, Perspective: cameraOriginValue(500, 0, 0), PerspectiveBase: base},
			{Offset: 1, Perspective: PerspectiveValue{
				Distance: 1000, OriginX: 100, OriginY: 50, HasDistance: true, HasOrigin: true,
			}, PerspectiveBase: base},
		},
		Timing: perspectiveTiming(),
	}

	value, ok := effect.samplePerspective(0)
	if !ok {
		t.Fatal("the effect reported no perspective")
	}
	if math.Abs(value.Distance-500) > perspectiveTolerance {
		t.Fatalf("0%% perspective = %v, want the explicit 500", value.Distance)
	}
	if math.Abs(value.OriginX) > perspectiveTolerance || math.Abs(value.OriginY) > perspectiveTolerance {
		t.Fatalf("0%% origin = %v,%v, want the explicit 0,0", value.OriginX, value.OriginY)
	}
}

// TestPerspectiveBasePadsAPartialOrigin checks the origin half of the implicit
// endpoint rule.
func TestPerspectiveBasePadsAPartialOrigin(t *testing.T) {
	base := cameraBase(800, true)
	effect := Effect{
		Frames: []Keyframe{{
			Offset: 0.5,
			Perspective: PerspectiveValue{
				OriginX: 100, OriginY: 50, HasOrigin: true,
			},
			PerspectiveBase: base,
		}},
		Timing: perspectiveTiming(),
	}

	start, ok := effect.samplePerspective(0)
	if !ok {
		t.Fatal("the partial effect reported no perspective")
	}
	if !start.HasOrigin || math.Abs(start.OriginX-7) > perspectiveTolerance ||
		math.Abs(start.OriginY-9) > perspectiveTolerance {
		t.Fatalf("0%% origin = %+v, want the static 7,9", start)
	}
	if start.HasDistance {
		t.Fatalf("an origin-only endpoint changed the distance: %+v", start)
	}
}
