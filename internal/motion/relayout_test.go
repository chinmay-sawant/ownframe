package motion

import (
	"strings"
	"testing"
	"time"

	"github.com/chinmay-sawant/blinkless/layout"
)

// layoutDisplay builds a two-box page whose boxes animate width, and nothing
// else. The display carries no operations: a layout player owns no paint
// target, which is the shape the queue has to handle.
func layoutDisplay(ids ...string) *Display {
	boxes := make([]layout.Box, len(ids))
	for i, id := range ids {
		boxes[i] = layout.Box{
			ID: id,
			Animations: []layout.Animation{{
				Name: "widen", DurationMS: 1000, IterationCount: 1, FillMode: "both",
			}},
		}
	}

	return &Display{
		Keyframes: []layout.KeyframesRule{{
			Name: "widen",
			Keyframes: []layout.Keyframe{
				{Offset: 0, Declarations: []layout.KeyframeDeclaration{
					{Property: "width", Value: "40px"},
				}},
				{Offset: 1, Declarations: []layout.KeyframeDeclaration{
					{Property: "width", Value: "120px"},
				}},
			},
		}},
		Boxes: boxes,
	}
}

// layoutRuntime binds a page at a fixed clock so the samples are deterministic.
func layoutRuntime(t *testing.T, ids ...string) (*Runtime, time.Time) {
	t.Helper()

	start := time.Unix(1000, 0)
	runtime := New()
	runtime.now = func() time.Time { return start }
	if err := runtime.BindAfter(nil, layoutDisplay(ids...)); err != nil {
		t.Fatal(err)
	}
	if got := runtime.Stats().LayoutBound; got != len(ids) {
		t.Fatalf("bound layout elements = %d, want %d", got, len(ids))
	}

	return runtime, start
}

// TestStepLayoutRunsOneElementPerFrame checks the budget and the element-id
// queue: one request per frame, the rest reported as waiting, and the deferred
// element sampling on the frame that runs it.
func TestStepLayoutRunsOneElementPerFrame(t *testing.T) {
	t.Parallel()

	runtime, start := layoutRuntime(t, "a", "b")

	requests := runtime.StepLayout(start.Add(500 * time.Millisecond))
	if len(requests) != 1 {
		t.Fatalf("requests = %+v, want one relayout", requests)
	}
	if requests[0].ElementID != "a" {
		t.Fatalf("first element = %q, want the id-ordered head a", requests[0].ElementID)
	}
	if got := requests[0].Overrides[0].Value; got != "80px" {
		t.Fatalf("width at 50%% = %q, want 80px", got)
	}
	stats := runtime.Stats()
	if stats.LayoutQueued != 1 || !stats.LayoutBudgetHit {
		t.Fatalf("queued = %d, hit = %v, want b waiting", stats.LayoutQueued, stats.LayoutBudgetHit)
	}

	// b was not sampled on the first frame, so its first request is its value
	// at this frame's clock.
	requests = runtime.StepLayout(start.Add(750 * time.Millisecond))
	if len(requests) != 1 || requests[0].ElementID != "b" {
		t.Fatalf("second frame = %+v, want b", requests)
	}
	if got := requests[0].Overrides[0].Value; got != "100px" {
		t.Fatalf("b width at the later frame = %q, want 100px", got)
	}

	// a comes back round after b, still live at the later clock.
	requests = runtime.StepLayout(start.Add(900 * time.Millisecond))
	if len(requests) != 1 || requests[0].ElementID != "a" {
		t.Fatalf("third frame = %+v, want a again", requests)
	}
}

// TestStepLayoutStopsAFinishedElement checks that a completed animation leaves
// the queue: the page stops relaying it out instead of replaying the end pose
// on every tick.
func TestStepLayoutStopsAFinishedElement(t *testing.T) {
	t.Parallel()

	runtime, start := layoutRuntime(t, "a")
	if requests := runtime.StepLayout(start.Add(2 * time.Second)); len(requests) != 1 {
		t.Fatalf("first sample = %+v, want the end pose", requests)
	}

	if requests := runtime.StepLayout(start.Add(3 * time.Second)); len(requests) != 0 {
		t.Fatalf("a finished element sampled again: %+v", requests)
	}
	if got := runtime.Stats().LayoutQueued; got != 0 {
		t.Fatalf("queued = %d after the animation finished, want 0", got)
	}
}

// TestReturnLayoutPutsTheSampleBack checks the retained-state debt path: a
// request the page cannot apply yet runs first on the next frame.
func TestReturnLayoutPutsTheSampleBack(t *testing.T) {
	t.Parallel()

	runtime, start := layoutRuntime(t, "a", "b")
	requests := runtime.StepLayout(start.Add(200 * time.Millisecond))
	if len(requests) != 1 || requests[0].ElementID != "a" {
		t.Fatalf("requests = %+v, want a", requests)
	}

	runtime.ReturnLayout(requests)
	if got := runtime.Stats().LayoutQueued; got != 2 {
		t.Fatalf("queued = %d, want both elements", got)
	}

	again := runtime.StepLayout(start.Add(300 * time.Millisecond))
	if len(again) != 1 || again[0].ElementID != "a" {
		t.Fatalf("returned sample = %+v, want a first", again)
	}
}

// TestStepLayoutSamplesEveryPropertyOnce checks the grouping rule: an element
// that animates width and height sends one request with one override each.
func TestStepLayoutSamplesEveryPropertyOnce(t *testing.T) {
	t.Parallel()

	display := layoutDisplay("card")
	display.Keyframes[0].Keyframes[0].Declarations = []layout.KeyframeDeclaration{
		{Property: "width", Value: "40px"},
		{Property: "height", Value: "20px"},
	}
	display.Keyframes[0].Keyframes[1].Declarations = []layout.KeyframeDeclaration{
		{Property: "width", Value: "120px"},
		{Property: "height", Value: "60px"},
	}

	start := time.Unix(2000, 0)
	runtime := New()
	runtime.now = func() time.Time { return start }
	if err := runtime.BindAfter(nil, display); err != nil {
		t.Fatal(err)
	}

	requests := runtime.StepLayout(start.Add(250 * time.Millisecond))
	if len(requests) != 1 {
		t.Fatalf("requests = %+v, want one element", requests)
	}
	if got := len(requests[0].Overrides); got != 2 {
		t.Fatalf("overrides = %d, want one per property", got)
	}
	want := map[string]string{"width": "60px", "height": "30px"}
	for _, override := range requests[0].Overrides {
		if want[override.Property] != override.Value {
			t.Fatalf("%s = %q, want %q", override.Property, override.Value, want[override.Property])
		}
	}
}

// TestStepLayoutPausedElementSamplesOnce keeps a frozen layout pose from
// dirtying the page every tick.
func TestStepLayoutPausedElementSamplesOnce(t *testing.T) {
	t.Parallel()

	display := layoutDisplay("card")
	display.Boxes[0].Animations[0].PlayState = "paused"
	start := time.Unix(3000, 0)
	runtime := New()
	runtime.now = func() time.Time { return start }
	if err := runtime.BindAfter(nil, display); err != nil {
		t.Fatal(err)
	}

	if requests := runtime.StepLayout(start.Add(100 * time.Millisecond)); len(requests) != 1 {
		t.Fatalf("first sample = %+v, want the frozen pose once", requests)
	}
	if requests := runtime.StepLayout(start.Add(200 * time.Millisecond)); len(requests) != 0 {
		t.Fatalf("paused element sampled again: %+v", requests)
	}
	if got := runtime.Stats().LayoutQueued; got != 0 {
		t.Fatalf("queued = %d after the paused sample, want 0", got)
	}
}

// TestStepLayoutReportsTheParseErrorOnce keeps a malformed layout value from
// being dropped silently.
func TestBadLayoutValueFailsTheBindOnce(t *testing.T) {
	t.Parallel()

	display := layoutDisplay("card")
	display.Keyframes[0].Keyframes[1].Declarations[0].Value = "banana"

	runtime := New()
	err := runtime.BindAfter(nil, display)
	if err == nil {
		t.Fatal("the bind accepted a malformed width")
	}
	if !strings.Contains(err.Error(), "invalid layout value") {
		t.Fatalf("error = %v, want the invalid layout value", err)
	}
	if got := runtime.Stats().LayoutBound; got != 0 {
		t.Fatalf("bound layout elements = %d, want the failed box left out", got)
	}
}
