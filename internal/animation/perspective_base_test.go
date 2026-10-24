package animation

import (
	"math"
	"testing"
)

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
