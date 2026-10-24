package motion

import (
	"strings"
	"testing"
	"time"

	"github.com/chinmay-sawant/blinkless/layout"
)

// geometryRecord is one linear transition record at the shared test duration.
func geometryRecord(property string) layout.Transition {
	return layout.Transition{Property: property, DurationMS: transitionFrameMS,
		TimingFunction: "linear"}
}

// geometryRow builds a page with one parent box and the target boxes below it.
// The parent carries no transitions; it gives top and left a non-zero origin so
// the test proves the inset is measured against it.
func geometryRow(parentX, parentY float64, targets ...layout.Box) *layout.Display {
	parent := layout.Box{ID: "row", Tag: "div", X: parentX, Y: parentY,
		W: testWidth, H: testHeight, ParentIndex: -1}
	for i := range targets {
		targets[i].ParentIndex = 0
	}
	boxes := append([]layout.Box{parent}, targets...)

	return &layout.Display{
		Width: testWidth, Height: testHeight, PointsPerPixel: 1, PixelPerPoint: 1,
		Boxes: boxes,
	}
}

// geometryTarget is one target box with a width and height and the transition
// records the property list should capture.
func geometryTarget(id string, x, y, w, h float64,
	records ...layout.Transition) layout.Box {
	return layout.Box{ID: id, Tag: "div", X: x, Y: y, W: w, H: h,
		Transitions: records}
}

// overrideValue returns the value of one property in a layout request.
func overrideValue(t *testing.T, request LayoutRequest, property string) string {
	t.Helper()
	for i := range request.Overrides {
		if request.Overrides[i].Property == property {
			return request.Overrides[i].Value
		}
	}
	t.Fatalf("%s: request %+v has no %s override", request.ElementID, request.Overrides, property)

	return ""
}

// TestWidthTransitionSamplesTheBoxThroughRelayout proves the geometry path: a
// width change is captured from the two boxes, the midpoint override is the
// interpolated length, and the final frame emits the destination once.
func TestWidthTransitionSamplesTheBoxThroughRelayout(t *testing.T) {
	t.Parallel()

	record := geometryRecord("width")
	old := geometryRow(0, 0, geometryTarget("target", 0, 0, 40, 20, record))
	next := geometryRow(0, 0, geometryTarget("target", 0, 0, 120, 20, record))

	r := New()
	start := testClock()
	r.now = func() time.Time { return start }
	captures := r.CaptureTransitions(old, next)
	if len(captures) != 1 {
		t.Fatalf("captures = %d, want the width change", len(captures))
	}
	r.BindTransitions(next, captures, start)
	if got := r.Stats().TransitionsBound; got != 1 {
		t.Fatalf("bound = %d, want the width run", got)
	}

	requests := r.StepLayout(start.Add(transitionFrameMS / 2 * time.Millisecond))
	if len(requests) != 1 || requests[0].ElementID != "target" {
		t.Fatalf("midpoint requests = %+v, want the target", requests)
	}
	if got := overrideValue(t, requests[0], "width"); got != "80px" {
		t.Fatalf("midpoint width = %q, want 80px", got)
	}

	requests = r.StepLayout(start.Add(transitionFrameMS * time.Millisecond))
	if len(requests) != 1 || overrideValue(t, requests[0], "width") != "120px" {
		t.Fatalf("final requests = %+v, want the destination width", requests)
	}
	if r.Active() {
		t.Fatal("a finished geometry transition still reported active")
	}
	if requests := r.StepLayout(start.Add(2 * transitionFrameMS * time.Millisecond)); len(requests) != 0 {
		t.Fatalf("a finished geometry transition relaid out again: %+v", requests)
	}
}

// TestHeightTransitionSamplesTheBox checks the second extent.
func TestHeightTransitionSamplesTheBox(t *testing.T) {
	t.Parallel()

	record := geometryRecord("height")
	old := geometryRow(0, 0, geometryTarget("target", 0, 0, 40, 20, record))
	next := geometryRow(0, 0, geometryTarget("target", 0, 0, 40, 60, record))

	r := New()
	start := testClock()
	r.now = func() time.Time { return start }
	r.BindTransitions(next, r.CaptureTransitions(old, next), start)

	requests := r.StepLayout(start.Add(transitionFrameMS / 2 * time.Millisecond))
	if len(requests) != 1 || overrideValue(t, requests[0], "height") != "40px" {
		t.Fatalf("midpoint requests = %+v, want height 40px", requests)
	}
}

// TestTopLeftTransitionInterpolatesAgainstTheParentOrigin checks that an inset
// is measured from the parent box, so a parent that does not start at the page
// origin still moves the element from its old box to its new one.
func TestTopLeftTransitionInterpolatesAgainstTheParentOrigin(t *testing.T) {
	t.Parallel()

	top, left := geometryRecord("top"), geometryRecord("left")
	old := geometryRow(10, 20, geometryTarget("target", 10, 20, 40, 20, top, left))
	next := geometryRow(10, 20, geometryTarget("target", 110, 70, 40, 20, top, left))

	r := New()
	start := testClock()
	r.now = func() time.Time { return start }
	captures := r.CaptureTransitions(old, next)
	if len(captures) != 2 {
		t.Fatalf("captures = %d, want top and left", len(captures))
	}
	r.BindTransitions(next, captures, start)

	requests := r.StepLayout(start.Add(transitionFrameMS / 2 * time.Millisecond))
	if len(requests) != 1 {
		t.Fatalf("requests = %+v, want one element", requests)
	}
	if got := overrideValue(t, requests[0], "top"); got != "25px" {
		t.Fatalf("midpoint top = %q, want 25px", got)
	}
	if got := overrideValue(t, requests[0], "left"); got != "50px" {
		t.Fatalf("midpoint left = %q, want 50px", got)
	}
}

// TestLayoutAnimationWinsOverWidthTransition checks the cascade: a keyframe on
// width owns the used value, so the transition capture is skipped and counted.
func TestLayoutAnimationWinsOverWidthTransition(t *testing.T) {
	t.Parallel()

	record := geometryRecord("width")
	old := geometryRow(0, 0, geometryTarget("target", 0, 0, 40, 20, record))
	target := geometryTarget("target", 0, 0, 40, 20, record)
	target.Animations = []layout.Animation{{
		Name: "widen", DurationMS: transitionFrameMS, IterationCount: 1,
		TimingFunction: "linear", FillMode: "both", PlayState: "running",
	}}
	next := geometryRow(0, 0, target)
	next.Keyframes = []layout.KeyframesRule{{
		Name: "widen",
		Keyframes: []layout.Keyframe{
			{Offset: 0, Declarations: []layout.KeyframeDeclaration{
				{Property: "width", Value: "40px"}}},
			{Offset: 1, Declarations: []layout.KeyframeDeclaration{
				{Property: "width", Value: "120px"}}},
		},
	}}

	r := New()
	if captures := r.CaptureTransitions(old, next); len(captures) != 0 {
		t.Fatalf("captures = %d, want the width transition left out", len(captures))
	}
	r.BindTransitions(next, nil, testClock())
	if got := r.Stats().TransitionsSkipped; got != 1 {
		t.Fatalf("skipped = %d, want the animation precedence counted", got)
	}
	if got := r.Stats().TransitionsBound; got != 0 {
		t.Fatalf("bound = %d, want no replacement", got)
	}
	if got := r.Stats().TransitionSkipReason; !strings.Contains(got, "animation") {
		t.Fatalf("skip reason = %q, want the animation precedence", got)
	}
}

// TestLayoutBudgetDefersTheSecondGeometryTransition proves the shared budget:
// the second width run waits, then samples at the clock of the frame that runs
// it rather than the frame it queued.
func TestLayoutBudgetDefersTheSecondGeometryTransition(t *testing.T) {
	t.Parallel()

	record := geometryRecord("width")
	old := geometryRow(0, 0,
		geometryTarget("a", 0, 0, 40, 20, record),
		geometryTarget("b", 0, 20, 40, 20, record))
	next := geometryRow(0, 0,
		geometryTarget("a", 0, 0, 100, 20, record),
		geometryTarget("b", 0, 20, 100, 20, record))

	r := New()
	start := testClock()
	r.now = func() time.Time { return start }
	captures := r.CaptureTransitions(old, next)
	if len(captures) != 2 {
		t.Fatalf("captures = %d, want both widths", len(captures))
	}
	r.BindTransitions(next, captures, start)

	requests := r.StepLayout(start.Add(transitionFrameMS / 2 * time.Millisecond))
	if len(requests) != 1 || requests[0].ElementID != "a" {
		t.Fatalf("first frame = %+v, want a only", requests)
	}
	stats := r.Stats()
	if stats.LayoutQueued != 1 || !stats.LayoutBudgetHit {
		t.Fatalf("queued = %d, hit = %v, want b deferred", stats.LayoutQueued, stats.LayoutBudgetHit)
	}

	requests = r.StepLayout(start.Add(3 * transitionFrameMS / 4 * time.Millisecond))
	if len(requests) != 1 || requests[0].ElementID != "b" {
		t.Fatalf("second frame = %+v, want b", requests)
	}
	if got := overrideValue(t, requests[0], "width"); got != "85px" {
		t.Fatalf("deferred width = %q, want the value at that frame (85px)", got)
	}
}

// TestDisplayExitIsSkippedWithAReason keeps the display cross-fade honest: the
// element leaves the new display, so there is no outgoing paint to animate and
// the capture reports the phase 04 layer instead of inventing one.
func TestDisplayExitIsSkippedWithAReason(t *testing.T) {
	t.Parallel()

	old := geometryRow(0, 0, geometryTarget("target", 0, 0, 40, 20,
		geometryRecord("display")))
	next := geometryRow(0, 0)

	r := New()
	if captures := r.CaptureTransitions(old, next); len(captures) != 0 {
		t.Fatalf("captures = %d, want none for a display exit", len(captures))
	}
	r.BindTransitions(next, nil, testClock())
	if got := r.Stats().TransitionsBlocked; got != 1 {
		t.Fatalf("blocked = %d, want the display exit counted", got)
	}
	if got := r.Stats().TransitionSkipReason; !strings.Contains(got, "display") {
		t.Fatalf("skip reason = %q, want the display exit", got)
	}
}

// stackDisplay builds two filled boxes and a paint order. The target's z-index
// and the order are the values the stacking rule flips.
func stackDisplay(zIndex int, order []int, records ...layout.Transition) *layout.Display {
	ops := []layout.DisplayOp{
		{ID: 1, Kind: layout.DisplayOpFillRect, X: 0, Y: 0, W: 40, H: 20,
			R: .2, G: .3, B: .4, Alpha: 1, ZIndex: zIndex, ZIndexSet: zIndex != 0},
		{ID: 2, Kind: layout.DisplayOpFillRect, X: 0, Y: 20, W: 40, H: 20,
			R: .5, G: .6, B: .7, Alpha: 1},
	}

	return &layout.Display{
		Width: testWidth, Height: testHeight, PointsPerPixel: 1, PixelPerPoint: 1,
		Ops: ops, OpOwners: []int{0, 1}, Order: order,
		Boxes: []layout.Box{
			{ID: "target", Tag: "div", X: 0, Y: 0, W: 40, H: 20, ParentIndex: -1,
				OperationIDs: []uint64{1}, Transitions: records},
			{ID: "next", Tag: "div", X: 0, Y: 20, W: 40, H: 20, ParentIndex: -1,
				OperationIDs: []uint64{2}},
		},
	}
}

// TestZIndexTransitionFlipsOrderAtTheMidpoint checks the integer rule: the old
// stacking stays until 50% of the eased progress, then the new order installs
// and the damage covers the flipped operations.
func TestZIndexTransitionFlipsOrderAtTheMidpoint(t *testing.T) {
	t.Parallel()

	record := layout.Transition{Property: "z-index", DurationMS: transitionFrameMS,
		Behavior: "allow-discrete", TimingFunction: "linear"}
	old := stackDisplay(0, []int{0, 1}, record)
	next := stackDisplay(5, []int{1, 0}, record)

	r := New()
	start := testClock()
	r.now = func() time.Time { return start }
	captures := r.CaptureTransitions(old, next)
	if len(captures) != 1 {
		t.Fatalf("captures = %d, want the z-index change", len(captures))
	}
	r.BindTransitions(next, captures, start)

	if got := next.Order; !sameOrder(got, []int{0, 1}) {
		t.Fatalf("order at the bind = %v, want the old 0,1", got)
	}
	if next.Ops[0].ZIndexSet || next.Ops[0].ZIndex != 0 {
		t.Fatalf("z-index at the bind = %d/%v, want the old unset 0", next.Ops[0].ZIndex,
			next.Ops[0].ZIndexSet)
	}

	if damage := r.Step(next, start.Add(40*time.Millisecond)); !damage.Empty() {
		t.Fatalf("a frame before the midpoint damaged %v", damage.Bounds())
	}
	if got := next.Order; !sameOrder(got, []int{0, 1}) {
		t.Fatalf("order before the midpoint = %v, want the old order", got)
	}

	damage := r.Step(next, start.Add(50*time.Millisecond))
	if damage.Empty() {
		t.Fatal("the midpoint flip damaged nothing")
	}
	if got := next.Order; !sameOrder(got, []int{1, 0}) {
		t.Fatalf("order at the midpoint = %v, want the new order", got)
	}
	if !next.Ops[0].ZIndexSet || next.Ops[0].ZIndex != 5 {
		t.Fatalf("z-index at the midpoint = %d/%v, want 5", next.Ops[0].ZIndex,
			next.Ops[0].ZIndexSet)
	}
	r.Step(next, start.Add(transitionFrameMS*time.Millisecond))
	if r.Active() {
		t.Fatal("a finished z-index transition still reported active")
	}
}

// sameOrder compares two paint orders.
func sameOrder(got, want []int) bool {
	if len(got) != len(want) {
		return false
	}
	for i := range got {
		if got[i] != want[i] {
			return false
		}
	}

	return true
}

// presenceDisplay builds one target box with a fill and an optional text
// operation. The visibility record travels on the target.
func presenceDisplay(text bool, alpha float64, records ...layout.Transition) *layout.Display {
	ops := []layout.DisplayOp{
		{ID: 1, Kind: layout.DisplayOpFillRect, X: 0, Y: 0, W: 40, H: 20,
			R: .2, G: .3, B: .4, Alpha: alpha},
	}
	ids := []uint64{1}
	if text {
		ops = append(ops, layout.DisplayOp{ID: 2, Kind: layout.DisplayOpText,
			X: 0, Y: 10, W: 20, H: 10, Text: "hi"})
		ids = append(ids, 2)
	}

	return &layout.Display{
		Width: testWidth, Height: testHeight, PointsPerPixel: 1, PixelPerPoint: 1,
		Ops: ops, OpOwners: []int{0, 0},
		Boxes: []layout.Box{{
			ID: "target", Tag: "div", X: 0, Y: 0, W: 40, H: 20, ParentIndex: -1,
			OperationIDs: ids, Transitions: records,
		}},
	}
}

// TestVisibilityStaysVisibleUntilTheEnd checks the leaving presence rule: the
// operation keeps the alpha on screen for the whole run and flips to the hidden
// endpoint only at 100%.
func TestVisibilityStaysVisibleUntilTheEnd(t *testing.T) {
	t.Parallel()

	record := layout.Transition{Property: "visibility", DurationMS: transitionFrameMS,
		Behavior: "allow-discrete", TimingFunction: "linear"}
	old := presenceDisplay(true, 1, record)
	next := presenceDisplay(false, 1, record)

	r := New()
	start := testClock()
	r.now = func() time.Time { return start }
	captures := r.CaptureTransitions(old, next)
	if len(captures) != 1 {
		t.Fatalf("captures = %d, want the visibility change", len(captures))
	}
	r.BindTransitions(next, captures, start)

	if got := next.Ops[0].Alpha; got != 1 {
		t.Fatalf("alpha at the bind = %v, want the visible 1", got)
	}
	if damage := r.Step(next, start.Add(25*time.Millisecond)); !damage.Empty() {
		t.Fatalf("a mid-run frame damaged %v, want the visible paint untouched", damage.Bounds())
	}
	if got := next.Ops[0].Alpha; got != 1 {
		t.Fatalf("alpha mid-run = %v, want the visible 1", got)
	}

	if damage := r.Step(next, start.Add(transitionFrameMS*time.Millisecond)); damage.Empty() {
		t.Fatal("the end flip damaged nothing")
	}
	if got := next.Ops[0].Alpha; got != presenceHiddenAlpha {
		t.Fatalf("alpha at the end = %v, want the hidden endpoint", got)
	}
	if r.Active() {
		t.Fatal("a finished visibility transition still reported active")
	}
}

// presenceHiddenAlpha mirrors the rule's hidden endpoint for the test.
const presenceHiddenAlpha = 1e-3

// TestVisibilityEntryFlipsAtTheStart checks the entering direction: the new
// operations are hidden at 0% and visible at the first interior sample.
func TestVisibilityEntryFlipsAtTheStart(t *testing.T) {
	t.Parallel()

	record := layout.Transition{Property: "visibility", DurationMS: transitionFrameMS,
		Behavior: "allow-discrete", TimingFunction: "linear"}
	old := presenceDisplay(false, 1, record)
	next := presenceDisplay(true, 1, record)

	r := New()
	start := testClock()
	r.now = func() time.Time { return start }
	r.BindTransitions(next, r.CaptureTransitions(old, next), start)
	for i := range next.Ops {
		if got := next.Ops[i].Alpha; got != presenceHiddenAlpha {
			t.Fatalf("op %d alpha at the bind = %v, want hidden", i, got)
		}
	}

	if damage := r.Step(next, start.Add(time.Millisecond)); damage.Empty() {
		t.Fatal("the start flip damaged nothing")
	}
	if got := next.Ops[0].Alpha; got != 1 {
		t.Fatalf("fill alpha after the start = %v, want visible 1", got)
	}
	if got := next.Ops[1].Alpha; got != 0 {
		t.Fatalf("text alpha after the start = %v, want the operation's opaque 0", got)
	}
}

// TestDiscreteWithoutAllowDiscreteDoesNotBind checks the gate: without
// allow-discrete the committed value snaps and no rule is built.
func TestDiscreteWithoutAllowDiscreteDoesNotBind(t *testing.T) {
	t.Parallel()

	record := layout.Transition{Property: "z-index", DurationMS: transitionFrameMS,
		TimingFunction: "linear"}
	old := stackDisplay(0, []int{0, 1}, record)
	next := stackDisplay(5, []int{1, 0}, record)

	r := New()
	captures := r.CaptureTransitions(old, next)
	if len(captures) != 0 {
		t.Fatalf("captures = %d, want none without allow-discrete", len(captures))
	}
	r.BindTransitions(next, captures, testClock())
	if got := r.Stats().TransitionsBound; got != 0 {
		t.Fatalf("bound = %d, want none", got)
	}
	if got := next.Order; !sameOrder(got, []int{1, 0}) {
		t.Fatalf("order = %v, want the committed new order", got)
	}
}

// positionDisplay builds one positioned operation at x. A position change
// moves the operation, which the paint renderer flips at the halfway point.
func positionDisplay(x float64, records ...layout.Transition) *layout.Display {
	op := layout.DisplayOp{ID: 1, Kind: layout.DisplayOpFillRect, X: x, Y: 0,
		W: 40, H: 20, R: .2, G: .3, B: .4, Alpha: 1, Positioned: true}

	return &layout.Display{
		Width: testWidth, Height: testHeight, PointsPerPixel: 1, PixelPerPoint: 1,
		Ops: []layout.DisplayOp{op}, OpOwners: []int{0},
		Boxes: []layout.Box{{
			ID: "target", Tag: "div", X: 0, Y: 0, W: 40, H: 20, ParentIndex: -1,
			OperationIDs: []uint64{1}, Transitions: records,
		}},
	}
}

// TestPositionTransitionKeepsTheHalfwayFlip wires the existing discrete paint
// renderer: a position change writes the old X/Y before 50% and the new X/Y at
// the midpoint.
func TestPositionTransitionKeepsTheHalfwayFlip(t *testing.T) {
	t.Parallel()

	record := layout.Transition{Property: "position", DurationMS: transitionFrameMS,
		Behavior: "allow-discrete", TimingFunction: "linear"}
	old := positionDisplay(0, record)
	next := positionDisplay(100, record)

	r := New()
	start := testClock()
	r.now = func() time.Time { return start }
	captures := r.CaptureTransitions(old, next)
	if len(captures) != 1 {
		t.Fatalf("captures = %d, want the position change", len(captures))
	}
	r.BindTransitions(next, captures, start)

	r.Step(next, start.Add(49*time.Millisecond))
	if got := next.Ops[0].X; got != 0 {
		t.Fatalf("X before the midpoint = %v, want the old 0", got)
	}
	r.Step(next, start.Add(50*time.Millisecond))
	if got := next.Ops[0].X; got != 100 {
		t.Fatalf("X at the midpoint = %v, want the new 100", got)
	}
}

// orderDisplay builds a flex-like pair in one of the two box orders. The target
// declares the order transition; the other box does not.
func orderDisplay(swapped bool, records ...layout.Transition) *layout.Display {
	first := layout.Box{ID: "a", Tag: "div", X: 0, Y: 0, W: 40, H: 20,
		ParentIndex: 0, Transitions: records}
	second := layout.Box{ID: "b", Tag: "div", X: 40, Y: 0, W: 40, H: 20, ParentIndex: 0}
	parent := layout.Box{ID: "row", Tag: "div", X: 0, Y: 0, W: 200, H: 40, ParentIndex: -1}
	if swapped {
		return &layout.Display{
			Width: testWidth, Height: testHeight, PointsPerPixel: 1, PixelPerPoint: 1,
			Boxes: []layout.Box{parent, second, first},
		}
	}

	return &layout.Display{
		Width: testWidth, Height: testHeight, PointsPerPixel: 1, PixelPerPoint: 1,
		Boxes: []layout.Box{parent, first, second},
	}
}

// TestOrderChangeIsReportedBlocked keeps the integer honest: the display
// changes the box order, but the engine's override path has no order property,
// so the change snaps and the reason names the missing path instead of faking a
// midpoint flip.
func TestOrderChangeIsReportedBlocked(t *testing.T) {
	t.Parallel()

	record := layout.Transition{Property: "order", DurationMS: transitionFrameMS,
		Behavior: "allow-discrete", TimingFunction: "linear"}
	old := orderDisplay(false, record)
	next := orderDisplay(true, record)

	r := New()
	if captures := r.CaptureTransitions(old, next); len(captures) != 0 {
		t.Fatalf("captures = %d, want no order renderer", len(captures))
	}
	r.BindTransitions(next, nil, testClock())
	if got := r.Stats().TransitionsBlocked; got != 1 {
		t.Fatalf("blocked = %d, want the order change counted", got)
	}
	if got := r.Stats().TransitionSkipReason; !strings.Contains(got, "order") {
		t.Fatalf("skip reason = %q, want the order property", got)
	}
}

// TestContentVisibilityFlipsAtTheMidpoint checks the discrete second presence
// property: content-visibility is a plain discrete value, so the remaining fill
// stays visible before 50% and hides at the boundary.
func TestContentVisibilityFlipsAtTheMidpoint(t *testing.T) {
	t.Parallel()

	record := layout.Transition{Property: "content-visibility", DurationMS: transitionFrameMS,
		Behavior: "allow-discrete", TimingFunction: "linear"}
	old := presenceDisplay(true, 1, record)
	next := presenceDisplay(false, 1, record)

	r := New()
	start := testClock()
	r.now = func() time.Time { return start }
	captures := r.CaptureTransitions(old, next)
	if len(captures) != 1 {
		t.Fatalf("captures = %d, want the content-visibility change", len(captures))
	}
	r.BindTransitions(next, captures, start)
	if got := next.Ops[0].Alpha; got != 1 {
		t.Fatalf("alpha at the bind = %v, want visible", got)
	}
	r.Step(next, start.Add(49*time.Millisecond))
	if got := next.Ops[0].Alpha; got != 1 {
		t.Fatalf("alpha before the midpoint = %v, want visible", got)
	}
	r.Step(next, start.Add(50*time.Millisecond))
	if got := next.Ops[0].Alpha; got != presenceHiddenAlpha {
		t.Fatalf("alpha at the midpoint = %v, want the hidden endpoint", got)
	}
}
