package animation

import (
	"math"
	"testing"
	"time"

	"github.com/chinmay-sawant/blinkless/layout"
	"github.com/chinmay-sawant/ownframe/internal/cssmotion"
)

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
	p, err := New([]Target{{Op: op, From: cssmotion.Read(op)}}, []Keyframe{
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
	p, err := New([]Target{{Op: op, From: cssmotion.Read(op)}}, []Keyframe{
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
			p, err := New([]Target{{Op: &layout.DisplayOp{}}}, test.frames, perspectiveTiming())
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
	p, err := New([]Target{{Op: op, From: cssmotion.Read(op)}}, []Keyframe{
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
	p, err := New([]Target{{Op: op, From: cssmotion.Read(op)}}, []Keyframe{
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
