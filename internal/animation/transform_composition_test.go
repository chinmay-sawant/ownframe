package animation

import (
	"math"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/chinmay-sawant/blinkless/layout"
	"github.com/chinmay-sawant/ownframe/internal/cssmotion"
)

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
	base := cssmotion.Read(op)
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
	base := cssmotion.Read(op)
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
			base := cssmotion.Read(op)
			// The element's own transform is scale(2), so the composed
			// animation matrix is visible as a factor of the world matrix.
			target := Target{Op: op, From: base, World: layout.Matrix2D{A: 2, D: 2}}
			frames := []Keyframe{
				{Offset: 0, Transform: scaleValue(1)},
				{Offset: 1, Transform: scaleValue(3)},
			}
			timing := Timing{Duration: time.Second, Iterations: 2, TimingFunction: "linear",
				Composition: test.mode}
			p, err := New([]Target{target}, frames, timing)
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
	base := cssmotion.Read(op)
	frames := []Keyframe{
		{Offset: 0, Transform: translateValue(0, 0), TimingFunction: "ease-in"},
		{Offset: 1, Transform: translateValue(100, 0), TimingFunction: "linear"},
	}
	p, err := New([]Target{{Op: op, From: base}}, frames,
		Timing{Duration: time.Second, Iterations: 1, TimingFunction: "linear"})
	if err != nil {
		t.Fatal(err)
	}
	start := time.Unix(92, 0)
	p.Start(start)
	p.Step(start.Add(500 * time.Millisecond))

	eased := cssmotion.Ease("ease-in", 0.5)
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
	base := cssmotion.Read(op)
	frames := []Keyframe{
		{Offset: 0, Transform: translateValue(0, 0), TimingFunction: "linear"},
		{Offset: 0.5, Transform: translateValue(40, 0), TimingFunction: "ease-in"},
		{Offset: 1, Transform: translateValue(80, 0), TimingFunction: "linear"},
	}
	p, err := New([]Target{{Op: op, From: base}}, frames,
		Timing{Duration: time.Second, Iterations: 1, TimingFunction: "linear"})
	if err != nil {
		t.Fatal(err)
	}
	start := time.Unix(93, 0)
	p.Start(start)
	p.Step(start.Add(750 * time.Millisecond))

	want := 40 + 40*cssmotion.Ease("ease-in", 0.5)
	if got := op.Transform().E; math.Abs(got-want) > matrixTolerance {
		t.Fatalf("second segment = %v, want %v from ease-in", got, want)
	}
}

// TestPerspectiveSegmentUsesKeyframeEasing checks the same rule on the
// perspective property.
func TestPerspectiveSegmentUsesKeyframeEasing(t *testing.T) {
	op := &layout.DisplayOp{Kind: layout.DisplayOpFillRect, Alpha: 1}
	base := cssmotion.Read(op)
	frames := []Keyframe{
		{Offset: 0, Perspective: cameraValue(500, true), TimingFunction: "ease-in"},
		{Offset: 1, Perspective: cameraValue(1000, true), TimingFunction: "linear"},
	}
	p, err := New([]Target{{Op: op, From: base}}, frames, perspectiveTiming())
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
	want := 500 + 500*cssmotion.Ease("ease-in", 0.5)
	if got := values[0].Perspective; math.Abs(got-want) > perspectiveTolerance {
		t.Fatalf("perspective midpoint = %v, want %v from the keyframe easing", got, want)
	}
}

// TestMixedCompositionKeepsDeclarationOrder checks that replace, add, and
// accumulate fold in list order. The rotate between the translates makes the
// order observable in the final translation.
func TestMixedCompositionKeepsDeclarationOrder(t *testing.T) {
	op := &layout.DisplayOp{Kind: layout.DisplayOpFillRect, Alpha: 1}
	base := cssmotion.Read(op)
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
	base := cssmotion.Read(op)
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
			base := cssmotion.Read(op)
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
			base = cssmotion.Read(op)
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
