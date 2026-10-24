package animation

import (
	"math"
	"testing"
	"time"

	"github.com/chinmay-sawant/blinkless/layout"
	"github.com/chinmay-sawant/ownframe/internal/cssmotion"
)

// threeDValue builds a sampled 3D list from its functions.
func threeDValue(functions ...TransformFunction) TransformValue {
	return TransformValue{Decomposed: IdentityTransform(), Functions: functions, ThreeD: true}
}

func TestTransformCSSSerializesThreeDimensionalFunctions(t *testing.T) {
	value := threeDValue(
		TransformFunction{Name: RotateY, Args: []float64{math.Pi / 4}},
		TransformFunction{Name: TranslateZ, Args: []float64{15}},
	)

	// 15 layout points are 20 CSS pixels, the unit the engine parses.
	if got, want := TransformCSS(value), "rotateY(45deg) translateZ(20px)"; got != want {
		t.Fatalf("TransformCSS = %q, want %q", got, want)
	}
}

func TestTransformCSSSerializesEveryFunctionUnit(t *testing.T) {
	tests := []struct {
		name string
		fn   TransformFunction
		want string
	}{
		{"translate", TransformFunction{Name: Translate, Args: []float64{7.5, 15}}, "translate(10px, 20px)"},
		{"translateX", TransformFunction{Name: TranslateX, Args: []float64{7.5}}, "translateX(10px)"},
		{"translateY", TransformFunction{Name: TranslateY, Args: []float64{15}}, "translateY(20px)"},
		{"translate3d", TransformFunction{Name: Translate3D, Args: []float64{7.5, 15, 22.5}},
			"translate3d(10px, 20px, 30px)"},
		{"translateZ", TransformFunction{Name: TranslateZ, Args: []float64{15}}, "translateZ(20px)"},
		{"scale", TransformFunction{Name: Scale, Args: []float64{2, 3}}, "scale(2, 3)"},
		{"scaleX", TransformFunction{Name: ScaleX, Args: []float64{2}}, "scaleX(2)"},
		{"scaleY", TransformFunction{Name: ScaleY, Args: []float64{3}}, "scaleY(3)"},
		{"scale3d", TransformFunction{Name: Scale3D, Args: []float64{2, 3, 4}}, "scale3d(2, 3, 4)"},
		{"scaleZ", TransformFunction{Name: ScaleZ, Args: []float64{4}}, "scaleZ(4)"},
		{"rotate", TransformFunction{Name: Rotate, Args: []float64{math.Pi / 2}}, "rotate(90deg)"},
		{"rotateX", TransformFunction{Name: RotateX, Args: []float64{math.Pi / 4}}, "rotateX(45deg)"},
		{"rotateY", TransformFunction{Name: RotateY, Args: []float64{math.Pi / 8}}, "rotateY(22.5deg)"},
		{"rotateZ", TransformFunction{Name: RotateZ, Args: []float64{math.Pi}}, "rotateZ(180deg)"},
		{"rotate3d", TransformFunction{Name: Rotate3D, Args: []float64{0, 1, 0, math.Pi / 2}},
			"rotate3d(0, 1, 0, 90deg)"},
		{"skew", TransformFunction{Name: Skew, Args: []float64{math.Pi / 4, 0}}, "skew(45deg, 0deg)"},
		{"skewX", TransformFunction{Name: SkewX, Args: []float64{math.Pi / 4}}, "skewX(45deg)"},
		{"skewY", TransformFunction{Name: SkewY, Args: []float64{0}}, "skewY(0deg)"},
		{"matrix", TransformFunction{Name: Matrix, Args: []float64{1, 0, 0, 1, 7.5, 15}},
			"matrix(1, 0, 0, 1, 10px, 20px)"},
		{"matrix3d", TransformFunction{Name: Matrix3D,
			Args: []float64{1, 0, 0, 0, 0, 1, 0, 0, 0, 0, 1, 0, 7.5, 15, 22.5, 1}},
			"matrix3d(1, 0, 0, 0, 0, 1, 0, 0, 0, 0, 1, 0, 10px, 20px, 30px, 1)"},
		{"perspective", TransformFunction{Name: Perspective, Args: []float64{375}}, "perspective(500px)"},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			if got := TransformCSS(threeDValue(test.fn)); got != test.want {
				t.Fatalf("TransformCSS = %q, want %q", got, test.want)
			}
		})
	}
}

func TestTransformCSSWritesATwoDimensionalValueAsAMatrix(t *testing.T) {
	value := TransformValue{
		Decomposed: Transform{Tx: 7.5, Ty: 15, Sx: 2, Sy: 3, Set: true},
	}

	// T(7.5,15) * S(2,3) in points, with the translation written in pixels.
	if got, want := TransformCSS(value), "matrix(2, 0, 0, 3, 10px, 20px)"; got != want {
		t.Fatalf("TransformCSS = %q, want %q", got, want)
	}
}

func TestTransformCSSWritesAnUnsetValueAsTheIdentity(t *testing.T) {
	if got, want := TransformCSS(TransformValue{}), "matrix(1, 0, 0, 1, 0px, 0px)"; got != want {
		t.Fatalf("TransformCSS = %q, want %q", got, want)
	}
}

func TestRestampValuesCarryTheSampledThreeDimensionalList(t *testing.T) {
	op := &layout.DisplayOp{Kind: layout.DisplayOpFillRect, Alpha: 1}
	p, err := New([]Target{{Op: op, From: cssmotion.Read(op)}},
		[]Keyframe{
			{Offset: 0, Transform: rotateValue(RotateY, 0, true)},
			{Offset: 1, Transform: rotateValue(RotateY, math.Pi/2, true)},
		},
		Timing{Duration: time.Second, Iterations: 1, TimingFunction: "linear"})
	if err != nil {
		t.Fatal(err)
	}
	start := time.Unix(70, 0)
	p.Start(start)
	p.Step(start.Add(500 * time.Millisecond))

	values := p.RestampValues()
	if len(values) != 1 {
		t.Fatalf("restamp values = %d, want one per target", len(values))
	}
	if values[0].Target != 0 || values[0].Transform != "rotateY(45deg)" {
		t.Fatalf("restamp value = %+v, want target 0 at %q", values[0], "rotateY(45deg)")
	}
}

func TestRestampValuesAreNilForATwoDimensionalSample(t *testing.T) {
	op := &layout.DisplayOp{Kind: layout.DisplayOpFillRect, Alpha: 1}
	p, err := New([]Target{{Op: op, From: cssmotion.Read(op)}},
		[]Keyframe{
			{Offset: 0, Transform: translateValue(0, 0)},
			{Offset: 1, Transform: translateValue(20, 10)},
		},
		Timing{Duration: time.Second, Iterations: 1, TimingFunction: "linear"})
	if err != nil {
		t.Fatal(err)
	}
	start := time.Unix(71, 0)
	p.Start(start)
	p.Step(start.Add(500 * time.Millisecond))

	if values := p.RestampValues(); values != nil {
		t.Fatalf("a 2D sample reported restamp values %+v", values)
	}
}

func TestRestampValuesFollowTheLastEffectOnAnElement(t *testing.T) {
	// A 3D effect followed by a 2D effect on the same element is a 2D frame:
	// the last effect wins, so the restamp list is empty.
	op := &layout.DisplayOp{Kind: layout.DisplayOpFillRect, Alpha: 1}
	target := Target{Op: op, From: cssmotion.Read(op)}
	threeD := Effect{
		Frames: []Keyframe{
			{Offset: 0, Transform: rotateValue(RotateY, 0, true)},
			{Offset: 1, Transform: rotateValue(RotateY, math.Pi, true)},
		},
		Timing: Timing{Duration: time.Second, Iterations: 1, TimingFunction: "linear"},
	}
	twoD := Effect{
		Frames: []Keyframe{
			{Offset: 0, Transform: translateValue(0, 0)},
			{Offset: 1, Transform: translateValue(10, 0)},
		},
		Timing: Timing{Duration: time.Second, Iterations: 1, TimingFunction: "linear"},
	}
	p, err := NewComposite([]Target{target}, []Effect{threeD, twoD})
	if err != nil {
		t.Fatal(err)
	}
	start := time.Unix(72, 0)
	p.Start(start)
	p.Step(start.Add(500 * time.Millisecond))

	if values := p.RestampValues(); values != nil {
		t.Fatalf("a 2D frame kept restamp values %+v", values)
	}
	if p.RestampPending() {
		t.Fatal("a 2D frame kept the restamp flag set")
	}
}

func TestRestampValuesSerializeATwoDimensionalSampleOnAThreeDElement(t *testing.T) {
	// A 2D keyframe on an element that already declares a 3D transform still
	// needs the engine pass, and the sampled matrix is the value it stamps.
	op := &layout.DisplayOp{Kind: layout.DisplayOpFillRect, Alpha: 1}
	target := Target{Op: op, From: cssmotion.Read(op), HasTransform3D: true}
	p, err := New([]Target{target},
		[]Keyframe{
			{Offset: 0, Transform: translateValue(0, 0)},
			{Offset: 1, Transform: translateValue(7.5, 0)},
		},
		Timing{Duration: time.Second, Iterations: 1, TimingFunction: "linear"})
	if err != nil {
		t.Fatal(err)
	}
	start := time.Unix(73, 0)
	p.Start(start)
	p.Step(start.Add(500 * time.Millisecond))

	values := p.RestampValues()
	if len(values) != 1 {
		t.Fatalf("restamp values = %d, want 1", len(values))
	}
	// Half of 7.5 points is 5 CSS pixels; the rest of the matrix is identity.
	if got, want := values[0].Transform, "matrix(1, 0, 0, 1, 5px, 0px)"; got != want {
		t.Fatalf("restamp transform = %q, want %q", got, want)
	}
}

func TestWantsRestampReadsTheKeyframesBeforeAnySample(t *testing.T) {
	threeD := []Keyframe{
		{Offset: 0, Transform: rotateValue(RotateY, 0, true)},
		{Offset: 1, Transform: rotateValue(RotateY, math.Pi, true)},
	}
	twoD := []Keyframe{
		{Offset: 0, Transform: translateValue(0, 0)},
		{Offset: 1, Transform: translateValue(10, 0)},
	}
	timing := Timing{Duration: time.Second}

	for _, test := range []struct {
		name    string
		target  Target
		frames  []Keyframe
		expects bool
	}{
		{name: "3D keyframe", target: Target{Op: &layout.DisplayOp{}}, frames: threeD, expects: true},
		{name: "2D keyframe", target: Target{Op: &layout.DisplayOp{}}, frames: twoD, expects: false},
		{name: "3D element", target: Target{Op: &layout.DisplayOp{}, HasTransform3D: true},
			frames: twoD, expects: true},
	} {
		t.Run(test.name, func(t *testing.T) {
			p, err := New([]Target{test.target}, test.frames, timing)
			if err != nil {
				t.Fatal(err)
			}
			if got := p.WantsRestamp(); got != test.expects {
				t.Fatalf("WantsRestamp = %v, want %v", got, test.expects)
			}
		})
	}
}
