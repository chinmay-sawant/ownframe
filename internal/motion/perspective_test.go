package motion

import (
	"math"
	"testing"
	"time"

	"github.com/chinmay-sawant/blinkless/layout"
)

// perspectiveDisplay replaces the fixture's translate keyframes with a
// perspective distance animation. PointsPerPixel is 1 in the fixture, so a
// 500px declaration is 500 points.
func perspectiveDisplay() *layout.Display {
	d := animatedDisplay("running", false)
	d.Keyframes[0].Keyframes[0].Declarations[0] = layout.KeyframeDeclaration{
		Property: "perspective", Value: "500px",
	}
	d.Keyframes[0].Keyframes[1].Declarations[0] = layout.KeyframeDeclaration{
		Property: "perspective", Value: "1000px",
	}

	return d
}

// addOrigin declares the same perspective-origin on both keyframes.
func addOrigin(d *layout.Display, value string) {
	for i := range d.Keyframes[0].Keyframes {
		d.Keyframes[0].Keyframes[i].Declarations = append(d.Keyframes[0].Keyframes[i].Declarations,
			layout.KeyframeDeclaration{Property: "perspective-origin", Value: value})
	}
}

func TestRuntimeCollectsAPerspectiveRequest(t *testing.T) {
	r := New()
	r.now = testClock
	d := perspectiveDisplay()

	if err := r.Bind(d); err != nil {
		t.Fatalf("bind: %v", err)
	}
	r.Step(d, testClock().Add(testDurationMS/2*time.Millisecond))

	requests := r.TakeRestamp()
	if len(requests) != 1 {
		t.Fatalf("restamp requests = %d, want one per element", len(requests))
	}
	request := requests[0]
	if request.ElementID != testChipID {
		t.Fatalf("request element = %q, want %q", request.ElementID, testChipID)
	}
	if !request.HasPerspective || math.Abs(request.Perspective-750) > 1e-9 {
		t.Fatalf("perspective = %+v, want the 750-point midpoint", request)
	}
	if request.Transform != "" {
		t.Fatalf("perspective-only request carried transform %q", request.Transform)
	}
	if request.HasPerspectiveOrigin {
		t.Fatalf("request reported an origin the keyframes did not declare: %+v", request)
	}
	// The keyframe parser resolves the 50%/50% default against the box, and the
	// driver lifts the box-relative offset into the page point the engine wants.
	if math.Abs(request.PerspectiveOriginX-testChipX-testChipSize/2) > 1e-9 ||
		math.Abs(request.PerspectiveOriginY-testChipY-testChipSize/2) > 1e-9 {
		t.Fatalf("default origin = %v,%v, want the box centre in page points",
			request.PerspectiveOriginX, request.PerspectiveOriginY)
	}
	if got := r.Stats().NeedsRestamp; got != 1 {
		t.Fatalf("restamp count = %d, want 1", got)
	}
}

func TestRuntimeCarriesAnOriginOnlyRequest(t *testing.T) {
	r := New()
	r.now = testClock
	d := animatedDisplay("running", false)
	d.Keyframes[0].Keyframes[0].Declarations[0] = layout.KeyframeDeclaration{
		Property: "perspective-origin", Value: "0 0",
	}
	d.Keyframes[0].Keyframes[1].Declarations[0] = layout.KeyframeDeclaration{
		Property: "perspective-origin", Value: "100% 100%",
	}

	if err := r.Bind(d); err != nil {
		t.Fatalf("bind: %v", err)
	}
	r.Step(d, testClock().Add(testDurationMS/2*time.Millisecond))

	requests := r.TakeRestamp()
	if len(requests) != 1 {
		t.Fatalf("restamp requests = %d, want 1", len(requests))
	}
	request := requests[0]
	if request.HasPerspective {
		t.Fatalf("an origin-only sample set HasPerspective: %+v", request)
	}
	if !request.HasPerspectiveOrigin {
		t.Fatalf("an origin-only sample lost the origin: %+v", request)
	}
	if request.PerspectiveOriginX != testChipX+testChipSize/2 ||
		request.PerspectiveOriginY != testChipY+testChipSize/2 {
		t.Fatalf("origin = %v,%v, want the page midpoint of the box",
			request.PerspectiveOriginX, request.PerspectiveOriginY)
	}
}

func TestRuntimeCarriesThePerspectiveOriginAsPageCoordinates(t *testing.T) {
	r := New()
	r.now = testClock
	d := perspectiveDisplay()
	addOrigin(d, "top left")

	if err := r.Bind(d); err != nil {
		t.Fatalf("bind: %v", err)
	}
	r.Step(d, testClock().Add(testDurationMS/2*time.Millisecond))

	requests := r.TakeRestamp()
	if len(requests) != 1 {
		t.Fatalf("restamp requests = %d, want 1", len(requests))
	}
	request := requests[0]
	if !request.HasPerspectiveOrigin {
		t.Fatalf("request lost the declared origin: %+v", request)
	}
	if request.PerspectiveOriginX != testChipX || request.PerspectiveOriginY != testChipY {
		t.Fatalf("origin = %v,%v, want the box top left %v,%v",
			request.PerspectiveOriginX, request.PerspectiveOriginY, testChipX, testChipY)
	}
}

func TestRuntimeSharesOneRequestForPerspectiveAndThreeD(t *testing.T) {
	r := New()
	r.now = testClock
	d := perspectiveDisplay()
	for i := range d.Keyframes[0].Keyframes {
		value := "rotateY(0deg)"
		if i == 1 {
			value = "rotateY(90deg)"
		}
		d.Keyframes[0].Keyframes[i].Declarations = append(d.Keyframes[0].Keyframes[i].Declarations,
			layout.KeyframeDeclaration{Property: "transform", Value: value})
	}

	if err := r.Bind(d); err != nil {
		t.Fatalf("bind: %v", err)
	}
	r.Step(d, testClock().Add(testDurationMS/2*time.Millisecond))

	requests := r.TakeRestamp()
	if len(requests) != 1 {
		t.Fatalf("restamp requests = %d, want one request for the element", len(requests))
	}
	if requests[0].Transform != "rotateY(45deg)" {
		t.Fatalf("transform = %q, want the sampled rotateY", requests[0].Transform)
	}
	if !requests[0].HasPerspective {
		t.Fatalf("the shared request lost the perspective: %+v", requests[0])
	}
}

func TestRuntimePerspectiveBindNeedsRetainedState(t *testing.T) {
	r := New()
	r.now = testClock

	if err := r.Bind(perspectiveDisplay()); err != nil {
		t.Fatalf("bind: %v", err)
	}
	if !r.NeedsRetainedState() {
		t.Fatal("a perspective keyframe did not ask for retained layout state")
	}

	r = New()
	r.now = testClock
	if err := r.Bind(animatedDisplay("running", false)); err != nil {
		t.Fatalf("bind 2D: %v", err)
	}
	if r.NeedsRetainedState() {
		t.Fatal("a 2D translate page asked for retained layout state")
	}
}

func TestRuntimePerspectiveOnlyPageCollectsNoMatrixWrite(t *testing.T) {
	r := New()
	r.now = testClock
	d := perspectiveDisplay()

	if err := r.Bind(d); err != nil {
		t.Fatalf("bind: %v", err)
	}
	r.Step(d, testClock().Add(testDurationMS/2*time.Millisecond))

	if d.Ops[0].XformSet {
		t.Fatal("a perspective-only sample wrote a 2D operation matrix")
	}
}
