package cssanimation

import (
	"errors"
	"strings"
	"testing"
	"time"

	"github.com/chinmay-sawant/blinkless/layout"
	"github.com/chinmay-sawant/ownframe/internal/animation"
)

// staticCamera is the perspective metadata a box with `perspective: 800px;
// perspective-origin: 10px 20px` carries. The test box is 200x100 CSS pixels at
// one point per pixel, so the resolved origin stays 10,20 points.
func staticCamera() layout.TransformInfo {
	return layout.TransformInfo{
		Parent:             layout.IdentityMatrix(),
		World:              layout.IdentityMatrix(),
		HasPerspective:     true,
		Perspective:        800,
		PerspectiveOriginX: 10,
		PerspectiveOriginY: 20,
	}
}

// cameraRule builds one running animation whose keyframes are the caller's.
func cameraRule(frames ...layout.Keyframe) []layout.KeyframesRule {
	return []layout.KeyframesRule{{Name: "camera", Keyframes: frames}}
}

// cameraDisplay builds one 200x100 target box with one fill and one animation.
func cameraDisplay(transform layout.TransformInfo, rules []layout.KeyframesRule) *layout.Display {
	return &layout.Display{
		Width: 200, Height: 100, PointsPerPixel: 1, PixelPerPoint: 1,
		Ops: []layout.DisplayOp{{ID: 1, Kind: layout.DisplayOpFillRect,
			X: 0, Y: 0, W: 200, H: 100, R: .1, G: .2, B: .3, Alpha: 1}},
		OpOwners: []int{0},
		Boxes: []layout.Box{{
			ID: "target", X: 0, Y: 0, W: 200, H: 100, ParentIndex: -1,
			OperationIDs: []uint64{1},
			Animations: []layout.Animation{{
				Name: "camera", DurationMS: 1000, IterationCount: 1,
				FillMode: "both", PlayState: "running",
			}},
			Transform: transform,
		}},
		Keyframes: rules,
	}
}

// sampleCamera binds the display and samples one 1 s animation at fraction.
func sampleCamera(t *testing.T, d *layout.Display, fraction float64) []animation.RestampValue {
	t.Helper()

	start := time.Unix(200, 0)
	bounds, _, err := BindBox(d, d.Boxes[0], 0, start)
	if err != nil {
		t.Fatalf("bind: %v", err)
	}
	if len(bounds) != 1 {
		t.Fatalf("players = %d, want one", len(bounds))
	}
	bounds[0].Player.Step(start.Add(time.Duration(fraction * float64(time.Second))))

	return bounds[0].Player.RestampValues()
}

// oneCameraSample binds and samples once, requiring one restamp request.
func oneCameraSample(t *testing.T, d *layout.Display, fraction float64) animation.RestampValue {
	t.Helper()

	values := sampleCamera(t, d, fraction)
	if len(values) != 1 {
		t.Fatalf("restamp values = %d, want one", len(values))
	}

	return values[0]
}

func TestParsePerspectiveOriginEdgeOffsets(t *testing.T) {
	size := perspectiveTestSize()

	for _, test := range []struct {
		name  string
		value string
		x, y  float64
	}{
		{name: "four values", value: "left 20px top 10px", x: 15, y: 7.5},
		{name: "four values from the far edges", value: "right 10% bottom 20px", x: 135, y: 60},
		{name: "four values vertical first", value: "bottom 20px right 10%", x: 135, y: 60},
		{name: "three values horizontal edge", value: "left 20px top", x: 15, y: 0},
		{name: "three values vertical edge", value: "right 10% bottom", x: 135, y: 75},
		{name: "three values vertical first", value: "top 10px left 20px", x: 15, y: 7.5},
		{name: "zero offsets", value: "left 0 top 0", x: 0, y: 0},
		{name: "percentage edge offset", value: "right 25% top", x: 112.5, y: 0},
		{name: "negative offset", value: "left -10px bottom 5px", x: -7.5, y: 71.25},
	} {
		t.Run(test.name, func(t *testing.T) {
			x, y, err := parsePerspectiveOrigin(test.value, size)
			if err != nil {
				t.Fatalf("parsePerspectiveOrigin(%q): %v", test.value, err)
			}
			if !near(x, test.x) || !near(y, test.y) {
				t.Fatalf("parsePerspectiveOrigin(%q) = %v,%v, want %v,%v",
					test.value, x, y, test.x, test.y)
			}
		})
	}
}

func TestParsePerspectiveOriginRejectsMalformedEdgeOffsets(t *testing.T) {
	size := perspectiveTestSize()

	for _, value := range []string{
		"left top 10px",        // offset must follow its edge
		"left 10px center",     // center takes no offset
		"center 10px top",      // center takes no offset
		"left 10px left 20px",  // both edges name the horizontal axis
		"left 10px 20px",       // missing the opposite edge
		"10px 20px 30px",       // a length is not an edge
		"left 10px top auto",   // auto is not an origin offset
		"left 10px top 20",     // unitless lengths are not CSS lengths
		"left 10px top 20px x", // too many components
	} {
		t.Run(value, func(t *testing.T) {
			if x, y, err := parsePerspectiveOrigin(value, size); err == nil {
				t.Fatalf("parsePerspectiveOrigin(%q) = %v,%v, want an error", value, x, y)
			}
		})
	}
}

// TestImplicitPerspectiveEndpointUsesTheStaticCamera is the regression for a
// partial keyframe list on an element that already has a camera: 0% holds the
// static 800px, not none.
func TestImplicitPerspectiveEndpointUsesTheStaticCamera(t *testing.T) {
	d := cameraDisplay(staticCamera(), cameraRule(layout.Keyframe{
		Offset: 0.5,
		Declarations: []layout.KeyframeDeclaration{
			{Property: "perspective", Value: "500px"},
		},
	}))

	start := oneCameraSample(t, d, 0)
	if !start.HasPerspective || !near(start.Perspective, 800) {
		t.Fatalf("0%% perspective = %+v, want the static 800", start)
	}

	middle := oneCameraSample(t, cameraDisplay(staticCamera(), d.Keyframes), 0.5)
	if !middle.HasPerspective || !near(middle.Perspective, 500) {
		t.Fatalf("50%% perspective = %+v, want the declared 500", middle)
	}
}

// TestImplicitPerspectiveOriginUsesTheStaticOrigin checks the origin half of
// the same rule: a partial origin list starts at the element's static point.
func TestImplicitPerspectiveOriginUsesTheStaticOrigin(t *testing.T) {
	d := cameraDisplay(staticCamera(), cameraRule(layout.Keyframe{
		Offset: 0.5,
		Declarations: []layout.KeyframeDeclaration{
			{Property: "perspective-origin", Value: "100px 50px"},
		},
	}))

	start := oneCameraSample(t, d, 0)
	if !start.HasPerspectiveOrigin || !near(start.PerspectiveOriginX, 10) ||
		!near(start.PerspectiveOriginY, 20) {
		t.Fatalf("0%% origin = %+v, want the static 10,20", start)
	}
	if start.HasPerspective {
		t.Fatalf("an origin-only endpoint changed the distance: %+v", start)
	}

	middle := oneCameraSample(t, cameraDisplay(staticCamera(), d.Keyframes), 0.5)
	if !near(middle.PerspectiveOriginX, 100) || !near(middle.PerspectiveOriginY, 50) {
		t.Fatalf("50%% origin = %v,%v, want the declared 100,50",
			middle.PerspectiveOriginX, middle.PerspectiveOriginY)
	}
}

// TestExplicitPerspectiveEndpointWins checks that a declared 0% or 100% stop
// is not replaced by the static base.
func TestExplicitPerspectiveEndpointWins(t *testing.T) {
	rules := cameraRule(
		layout.Keyframe{Offset: 0, Declarations: []layout.KeyframeDeclaration{
			{Property: "perspective", Value: "500px"},
			{Property: "perspective-origin", Value: "0px 0px"},
		}},
		layout.Keyframe{Offset: 1, Declarations: []layout.KeyframeDeclaration{
			{Property: "perspective", Value: "1000px"},
			{Property: "perspective-origin", Value: "100px 50px"},
		}},
	)

	start := oneCameraSample(t, cameraDisplay(staticCamera(), rules), 0)
	if !near(start.Perspective, 500) {
		t.Fatalf("0%% perspective = %v, want the explicit 500", start.Perspective)
	}
	if !start.HasPerspectiveOrigin || !near(start.PerspectiveOriginX, 0) ||
		!near(start.PerspectiveOriginY, 0) {
		t.Fatalf("0%% origin = %+v, want the explicit 0,0", start)
	}

	end := oneCameraSample(t, cameraDisplay(staticCamera(), rules), 1)
	if !near(end.Perspective, 1000) {
		t.Fatalf("100%% perspective = %v, want the explicit 1000", end.Perspective)
	}
	if !near(end.PerspectiveOriginX, 100) || !near(end.PerspectiveOriginY, 50) {
		t.Fatalf("100%% origin = %v,%v, want the explicit 100,50",
			end.PerspectiveOriginX, end.PerspectiveOriginY)
	}
}

// TestStaticNoneImplicitPerspectiveStaysNone pins that a missing camera is not
// invented from the box default.
func TestStaticNoneImplicitPerspectiveStaysNone(t *testing.T) {
	none := layout.TransformInfo{Parent: layout.IdentityMatrix(), World: layout.IdentityMatrix()}
	d := cameraDisplay(none, cameraRule(layout.Keyframe{
		Offset: 0.5,
		Declarations: []layout.KeyframeDeclaration{
			{Property: "perspective", Value: "500px"},
		},
	}))

	start := oneCameraSample(t, d, 0)
	if !start.HasPerspective || !near(start.Perspective, 0) {
		t.Fatalf("0%% perspective = %+v, want a none endpoint", start)
	}

	// Before the midpoint CSS is discrete across none.
	early := oneCameraSample(t, cameraDisplay(none, d.Keyframes), 0.1)
	if !early.HasPerspective || !near(early.Perspective, 0) {
		t.Fatalf("10%% perspective = %+v, want the none endpoint", early)
	}
}

// TestMalformedPerspectiveOriginErrorsOnce pins that a bad declaration is a
// single bind error, not one per keyframe.
func TestMalformedPerspectiveOriginErrorsOnce(t *testing.T) {
	d := cameraDisplay(staticCamera(), cameraRule(
		layout.Keyframe{Offset: 0, Declarations: []layout.KeyframeDeclaration{
			{Property: "perspective-origin", Value: "left 10px center"},
		}},
		layout.Keyframe{Offset: 1, Declarations: []layout.KeyframeDeclaration{
			{Property: "perspective-origin", Value: "left 10px center"},
		}},
	))

	_, _, err := BindBox(d, d.Boxes[0], 0, time.Unix(200, 0))
	if err == nil {
		t.Fatal("a malformed perspective-origin bound without an error")
	}
	if !errors.Is(err, ErrPerspective) {
		t.Fatalf("error = %v, want ErrPerspective", err)
	}
	if got := strings.Count(err.Error(), "invalid perspective value"); got != 1 {
		t.Fatalf("error carries the perspective failure %d times, want one: %v", got, err)
	}
}
