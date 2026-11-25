package motion

import (
	"github.com/chinmay-sawant/blinkless/layout"
	"math"
	"strings"
	"testing"
	"time"
)

const (
	transitionChipID  = "chip"
	transitionOpID    = 31
	transitionFrameMS = 100
)

// transitionChip builds the smallest display a transition needs: one fill
// operation owned by one box, with an optional element opacity and baked
// translate. The records are the transition declarations the new box carries.
func transitionChip(rgb, opacity, translate float64, records ...layout.Transition) *layout.Display {
	op := layout.DisplayOp{
		ID: transitionOpID, Kind: layout.DisplayOpFillRect,
		X: 10, Y: 10, W: 20, H: 20,
		R: rgb, G: rgb, B: rgb, Alpha: 1,
	}
	op.SetPaintOpacity(opacity)
	if translate != 0 {
		op.SetXform(layout.Matrix2D{A: 1, D: 1, E: translate})
	}
	return &layout.Display{
		Width: testWidth, Height: testHeight, PointsPerPixel: 1, PixelPerPoint: 1,
		Ops:      []layout.DisplayOp{op},
		OpOwners: []int{0},
		Boxes: []layout.Box{{
			ID: transitionChipID, Tag: "div",
			X: 10, Y: 10, W: 20, H: 20, ParentIndex: -1,
			OperationIDs: []uint64{transitionOpID},
			Transitions:  records,
		}},
	}
}

// linearTransition is one paint transition record with the test duration.
func linearTransition(property string) layout.Transition {
	return layout.Transition{Property: property, DurationMS: transitionFrameMS,
		TimingFunction: "linear"}
}

// near reports one float comparison at the paint epsilon.
func near(got, want float64) bool { return math.Abs(got-want) <= transitionPaintEpsilon }

// TestCaptureBindsColourOpacityAndTransform covers the phase gate: a class
// change on a colour, opacity, and transform starts all three renderers from
// one commit, and the midpoint writes each field without application code.
func TestCaptureBindsColourOpacityAndTransform(t *testing.T) {
	records := []layout.Transition{
		linearTransition("background-color"),
		linearTransition("opacity"),
		linearTransition("transform"),
	}
	old := transitionChip(0, 1, 0, records...)
	next := transitionChip(1, .5, 10, records...)

	r := New()
	start := testClock()
	r.now = func() time.Time { return start }

	captures := r.CaptureTransitions(old, next)
	if len(captures) != 3 {
		t.Fatalf("captures = %d, want 3", len(captures))
	}
	r.BindTransitions(next, captures, start)
	if got := r.Stats().TransitionsBound; got != 3 {
		t.Fatalf("bound transitions = %d, want 3", got)
	}
	if !r.Active() {
		t.Fatal("a newly bound transition reported idle")
	}

	damage := r.Step(next, start.Add(transitionFrameMS/2*time.Millisecond))
	if damage.Empty() {
		t.Fatal("a transition midpoint produced no damage")
	}
	op := &next.Ops[0]
	if !near(op.R, .5) || !near(op.G, .5) || !near(op.B, .5) {
		t.Fatalf("midpoint colour = %v,%v,%v, want .5", op.R, op.G, op.B)
	}
	if got := op.Opacity(); !near(got, .75) {
		t.Fatalf("midpoint opacity = %v, want .75", got)
	}
	if !op.XformSet || !near(op.Transform().E, 5) {
		t.Fatalf("midpoint matrix = %+v, want translate 5", op.Transform())
	}
	if op.X != 10 {
		t.Fatalf("a transform transition moved X to %v, want the layout 10", op.X)
	}
	if got := r.Stats().TransitionsActive; got != 3 {
		t.Fatalf("active transitions = %d, want 3", got)
	}

	r.Step(next, start.Add(transitionFrameMS*time.Millisecond))
	if !near(op.R, 1) || !near(op.Opacity(), .5) || !near(op.Transform().E, 10) {
		t.Fatalf("final paint = colour %v opacity %v matrix %+v",
			op.R, op.Opacity(), op.Transform())
	}
	if r.Active() {
		t.Fatal("a finished transition still reported active")
	}
	if _, ok := r.NextFrame(start.Add(transitionFrameMS * time.Millisecond)); ok {
		t.Fatal("a finished transition asked for another frame")
	}
	if damage := r.Step(next, start.Add(transitionFrameMS*time.Millisecond)); !damage.Empty() {
		t.Fatalf("a finished transition damaged %v", damage.Bounds())
	}
}

// TestInterruptedTransitionReversesFromThePresentedPaint checks the from-value
// rule: the second change captures the operation the first renderer has been
// writing, so the reversal starts at the paint on screen.
func TestInterruptedTransitionReversesFromThePresentedPaint(t *testing.T) {
	record := layout.Transition{Property: "background-color", DurationMS: 1000,
		TimingFunction: "linear"}
	old := transitionChip(0, 1, 0, record)
	middle := transitionChip(1, 1, 0, record)

	r := New()
	start := testClock()
	r.now = func() time.Time { return start }
	captures := r.CaptureTransitions(old, middle)
	r.BindTransitions(middle, captures, start)

	quarter := time.Second / 4
	r.Step(middle, start.Add(quarter))
	presented := middle.Ops[0].R
	if !near(presented, .25) {
		t.Fatalf("presented paint = %v, want .25", presented)
	}

	// The class goes back to the base colour while the first run is still in
	// flight. The new capture must read the presented paint as its from paint,
	// and the reverse is shortened by the progress already spent, so sample at
	// the bind instant before a later wall-clock tick moves it.
	next := transitionChip(0, 1, 0, record)
	captures = r.CaptureTransitions(middle, next)
	if len(captures) != 1 {
		t.Fatalf("interrupt captures = %d, want 1", len(captures))
	}
	r.BindTransitions(next, captures, start.Add(quarter))
	r.Step(next, start.Add(quarter))
	if got := next.Ops[0].R; !near(got, presented) {
		t.Fatalf("reversal start = %v, want the presented %v", got, presented)
	}

	// The reverse is quarter-length, so it finishes well before the declared
	// one-second duration.
	r.Step(next, start.Add(quarter+quarter))
	if got := next.Ops[0].R; !near(got, 0) {
		t.Fatalf("reversal end = %v, want 0 at the shortened deadline", got)
	}
}

// TestAllExpandsToChangedPaintProperties checks that transition-property: all
// binds only the properties whose selected paint actually changed. A record
// with no change produces no renderer and no frame request.
func TestAllExpandsToChangedPaintProperties(t *testing.T) {
	record := layout.Transition{Property: "all", DurationMS: transitionFrameMS,
		TimingFunction: "linear"}
	old := transitionChip(0, 1, 0, record)
	next := transitionChip(1, .5, 0, record)

	r := New()
	start := testClock()
	r.now = func() time.Time { return start }
	captures := r.CaptureTransitions(old, next)
	if len(captures) != 2 {
		t.Fatalf("captures = %d, want the colour and opacity only", len(captures))
	}
	r.BindTransitions(next, captures, start)
	if got := r.Stats().TransitionsBound; got != 2 {
		t.Fatalf("bound transitions = %d, want 2", got)
	}
	if next.Ops[0].XformSet {
		t.Fatal("an unchanged transform bound a matrix transition")
	}
}

// TestUnrelatedRebindKeepsTheRunningTransition checks the registry handoff: a
// publish that recomputes the same destination does not restart the clock, and
// the renderer only moves to the new operations.
func TestUnrelatedRebindKeepsTheRunningTransition(t *testing.T) {
	record := layout.Transition{Property: "background-color", DurationMS: 1000,
		TimingFunction: "linear"}
	old := transitionChip(0, 1, 0, record)
	middle := transitionChip(1, 1, 0, record)

	r := New()
	start := testClock()
	r.now = func() time.Time { return start }
	captures := r.CaptureTransitions(old, middle)
	r.BindTransitions(middle, captures, start)
	r.Step(middle, start.Add(250*time.Millisecond))
	if got := middle.Ops[0].R; !near(got, .25) {
		t.Fatalf("presented paint = %v, want .25", got)
	}

	// An unrelated redraw publishes the same stylesheet. The static base is
	// the destination again, so the capture must not see a new style change.
	again := transitionChip(1, 1, 0, record)
	rebindAt := start.Add(250 * time.Millisecond)
	if got := r.CaptureTransitions(middle, again); len(got) != 0 {
		t.Fatalf("unrelated rebind captured %d changes, want 0", len(got))
	}
	r.BindTransitions(again, nil, rebindAt)
	if got := r.Stats().TransitionsStarted; got != 0 {
		t.Fatalf("unrelated rebind started %d transitions, want 0", got)
	}
	if got := r.Stats().TransitionsBound; got != 1 {
		t.Fatalf("bound transitions after rebind = %d, want 1", got)
	}

	r.Step(again, start.Add(500*time.Millisecond))
	if got := again.Ops[0].R; !near(got, .5) {
		t.Fatalf("rebound sample = %v, want .5 from the original clock", got)
	}
}

// TestTransformTransitionInterpolatesARotation checks the matrix path beyond a
// pure translation: the midpoint of a 90 degree turn is the 45 degree matrix,
// which a component-wise A-F lerp would not produce.
func TestTransformTransitionInterpolatesARotation(t *testing.T) {
	record := linearTransition("transform")
	old := transitionChip(1, 1, 0, record)
	next := transitionChip(1, 1, 0, record)
	next.Ops[0].SetXform(layout.Matrix2D{
		A: math.Cos(math.Pi / 2), B: math.Sin(math.Pi / 2),
		C: -math.Sin(math.Pi / 2), D: math.Cos(math.Pi / 2),
	})

	r := New()
	start := testClock()
	r.now = func() time.Time { return start }
	captures := r.CaptureTransitions(old, next)
	if len(captures) != 1 {
		t.Fatalf("captures = %d, want 1", len(captures))
	}
	r.BindTransitions(next, captures, start)
	r.Step(next, start.Add(transitionFrameMS/2*time.Millisecond))

	got := next.Ops[0].Transform()
	if !near(got.A, math.Cos(math.Pi/4)) || !near(got.B, math.Sin(math.Pi/4)) {
		t.Fatalf("midpoint matrix = %+v, want a 45 degree rotation", got)
	}
}

// TestAnimationWinsOnTheSameOpacityField checks the step order: the transition
// writes first, the animation writes second, and the animation's value is what
// the frame shows.
func TestAnimationWinsOnTheSameOpacityField(t *testing.T) {
	record := linearTransition("opacity")
	old := transitionChip(1, 1, 0, record)
	next := transitionChip(1, .5, 0, record)
	next.Boxes[0].Animations = []layout.Animation{{
		Name: "fade", DurationMS: transitionFrameMS, IterationCount: 1,
		TimingFunction: "linear", FillMode: "both", PlayState: "running",
	}}
	next.Keyframes = []layout.KeyframesRule{{
		Name: "fade",
		Keyframes: []layout.Keyframe{
			{Offset: 0, Declarations: []layout.KeyframeDeclaration{
				{Property: "opacity", Value: "0"}}},
			{Offset: 1, Declarations: []layout.KeyframeDeclaration{
				{Property: "opacity", Value: "1"}}},
		},
	}}

	r := New()
	start := testClock()
	r.now = func() time.Time { return start }
	if err := r.Bind(next); err != nil {
		t.Fatalf("bind animation: %v", err)
	}
	captures := r.CaptureTransitions(old, next)
	if len(captures) != 1 {
		t.Fatalf("captures = %d, want 1", len(captures))
	}
	r.BindTransitions(next, captures, start)

	half := start.Add(transitionFrameMS / 2 * time.Millisecond)
	r.Step(next, half)
	if got := next.Ops[0].Opacity(); !near(got, .5) {
		t.Fatalf("opacity at the midpoint = %v, want the animation's .5, not the transition's .75", got)
	}
}

// TestTransitionFitFailureSnapsAndCounts covers the same-op-count rule: a
// changed operation count or kind never half-interpolates, and the skip is
// reported.
func TestTransitionFitFailureSnapsAndCounts(t *testing.T) {
	t.Run("count", func(t *testing.T) {
		record := linearTransition("background-color")
		old := transitionChip(0, 1, 0, record)
		next := transitionChip(1, 1, 0, record)
		second := layout.DisplayOp{ID: transitionOpID + 1, Kind: layout.DisplayOpFillRect,
			X: 40, Y: 10, W: 20, H: 20, R: 1, Alpha: 1}
		next.Ops = append(next.Ops, second)
		next.OpOwners = append(next.OpOwners, 0)
		next.Boxes[0].OperationIDs = append(next.Boxes[0].OperationIDs, second.ID)

		r := New()
		captures := r.CaptureTransitions(old, next)
		if len(captures) != 0 {
			t.Fatalf("captures = %d, want 0 on a changed op count", len(captures))
		}
		r.BindTransitions(next, captures, testClock())
		if got := r.Stats().TransitionsSkipped; got != 1 {
			t.Fatalf("skipped = %d, want 1", got)
		}
		if got := r.Stats().TransitionsBound; got != 0 {
			t.Fatalf("bound = %d, want 0", got)
		}
		if next.Ops[0].R != 1 {
			t.Fatalf("the new paint was not left in place: %v", next.Ops[0].R)
		}
	})

	t.Run("kind", func(t *testing.T) {
		record := linearTransition("opacity")
		old := transitionChip(1, 1, 0, record)
		next := transitionChip(1, .5, 0, record)
		next.Ops[0].Kind = layout.DisplayOpText

		r := New()
		captures := r.CaptureTransitions(old, next)
		if len(captures) != 0 {
			t.Fatalf("captures = %d, want 0 on a changed op kind", len(captures))
		}
		r.BindTransitions(next, captures, testClock())
		if got := r.Stats().TransitionsSkipped; got != 1 {
			t.Fatalf("skipped = %d, want 1", got)
		}
	})
}

// TestDelayedTransitionAsksForTheDelayEnd checks pacing: a delayed run holds
// its from paint without rewriting it, and the runtime asks for the delay end
// rather than the next frame.
func TestDelayedTransitionAsksForTheDelayEnd(t *testing.T) {
	record := layout.Transition{Property: "background-color", DurationMS: transitionFrameMS,
		DelayMS: transitionFrameMS, TimingFunction: "linear"}
	old := transitionChip(0, 1, 0, record)
	next := transitionChip(1, 1, 0, record)

	r := New()
	start := testClock()
	r.now = func() time.Time { return start }
	captures := r.CaptureTransitions(old, next)
	r.BindTransitions(next, captures, start)

	if at, ok := r.NextFrame(start); !ok || !at.Equal(start) {
		t.Fatalf("first wake = %v (%v), want the bind time", at, ok)
	}
	r.Step(next, start)
	if got := next.Ops[0].R; !near(got, 0) {
		t.Fatalf("delayed from paint = %v, want 0", got)
	}
	wake, ok := r.NextFrame(start)
	if !ok || !wake.Equal(start.Add(transitionFrameMS*time.Millisecond)) {
		t.Fatalf("delayed wake = %v (%v), want the delay end", wake, ok)
	}

	// A frame that arrives before the delay end must not rewrite the same
	// paint or move the deadline.
	damage := r.Step(next, start.Add(transitionFrameMS/2*time.Millisecond))
	if !damage.Empty() {
		t.Fatalf("a frame inside the delay damaged %v", damage.Bounds())
	}
	wake, ok = r.NextFrame(start.Add(transitionFrameMS / 2 * time.Millisecond))
	if !ok || !wake.Equal(start.Add(transitionFrameMS*time.Millisecond)) {
		t.Fatalf("mid-delay wake = %v (%v), want the delay end", wake, ok)
	}

	r.Step(next, start.Add(transitionFrameMS*time.Millisecond))
	if got := next.Ops[0].R; !near(got, 0) {
		t.Fatalf("delay-end paint = %v, want the from value", got)
	}
	r.Step(next, start.Add(transitionFrameMS*2*time.Millisecond))
	if got := next.Ops[0].R; !near(got, 1) {
		t.Fatalf("final paint = %v, want 1", got)
	}
	if _, ok := r.NextFrame(start.Add(transitionFrameMS * 2 * time.Millisecond)); ok {
		t.Fatal("a finished transition still asked for a frame")
	}
}

// TestReleaseTransitionsFreezesAtThePresentedPaint checks the switch rule: the
// renderer stops writing and the operation keeps the sampled value.
func TestReleaseTransitionsFreezesAtThePresentedPaint(t *testing.T) {
	record := linearTransition("background-color")
	old := transitionChip(0, 1, 0, record)
	next := transitionChip(1, 1, 0, record)

	r := New()
	start := testClock()
	r.now = func() time.Time { return start }
	captures := r.CaptureTransitions(old, next)
	r.BindTransitions(next, captures, start)
	r.Step(next, start.Add(transitionFrameMS/2*time.Millisecond))
	presented := next.Ops[0].R

	r.ReleaseTransitions()
	if r.Active() {
		t.Fatal("a released transition still reported active")
	}
	if damage := r.Step(next, start.Add(transitionFrameMS*time.Millisecond)); !damage.Empty() {
		t.Fatalf("a released transition damaged %v", damage.Bounds())
	}
	if got := next.Ops[0].R; !near(got, presented) {
		t.Fatalf("released paint = %v, want the presented %v", got, presented)
	}
}

// TestCaptureTransitionsFiltersUntransitionedBoxes verifies that indexBoxes only
// indexes boxes whose IDs declare transitions in either previous or next display.
func TestCaptureTransitionsFiltersUntransitionedBoxes(t *testing.T) {
	record := linearTransition("background-color")
	old := transitionChip(0, 1, 0, record)
	next := transitionChip(1, 1, 0, record)

	for i := 0; i < 50; i++ {
		id := "unrelated-" + string(rune('a'+i))
		old.Boxes = append(old.Boxes, layout.Box{ID: id, Tag: "div"})
		next.Boxes = append(next.Boxes, layout.Box{ID: id, Tag: "div"})
	}

	r := New()
	captures := r.CaptureTransitions(old, next)
	if len(captures) != 1 {
		t.Fatalf("captures = %d, want 1", len(captures))
	}
	if captures[0].key.element.id != transitionChipID {
		t.Fatalf("captured element = %s, want %s", captures[0].key.element.id, transitionChipID)
	}
}

// TestCaptureTransitionsNoTransitionsReturnsNil checks that CaptureTransitions
// returns nil immediately without allocating box maps when no transitions are declared.
func TestCaptureTransitionsNoTransitionsReturnsNil(t *testing.T) {
	old := &layout.Display{
		Boxes: []layout.Box{{ID: "b1", Tag: "div"}, {ID: "b2", Tag: "div"}},
	}
	next := &layout.Display{
		Boxes: []layout.Box{{ID: "b1", Tag: "div"}, {ID: "b2", Tag: "div"}},
	}
	r := New()
	if captures := r.CaptureTransitions(old, next); captures != nil {
		t.Fatalf("captures = %v, want nil", captures)
	}
}

// startRunningTransition starts a one-second colour run and advances it to its
// midpoint. The caller then cancels it with one of the rules below.
func startRunningTransition(t *testing.T) (*Runtime, time.Time, *layout.Display) {
	t.Helper()

	record := layout.Transition{Property: "background-color", DurationMS: 1000,
		TimingFunction: "linear"}
	old := transitionChip(0, 1, 0, record)
	middle := transitionChip(1, 1, 0, record)
	r := New()
	start := testClock()
	r.now = func() time.Time { return start }
	r.BindTransitions(middle, r.CaptureTransitions(old, middle), start)
	r.Step(middle, start.Add(500*time.Millisecond))
	if got := middle.Ops[0].R; !near(got, .5) {
		t.Fatalf("presented paint = %v, want .5", got)
	}

	return r, start, middle
}

// assertCancelled paints the final value once and never again. The committed
// display already carries it; the runtime must not write the cancelled run's
// stale destination over it on this or any later frame.
func assertCancelled(t *testing.T, r *Runtime, d *layout.Display, at time.Time) {
	t.Helper()

	if got := r.Stats().TransitionsBound; got != 0 {
		t.Fatalf("bound after cancel = %d, want 0", got)
	}
	if got := r.Stats().TransitionsCancelled; got != 1 {
		t.Fatalf("cancelled = %d, want 1", got)
	}
	if got := d.Ops[0].R; !near(got, .25) {
		t.Fatalf("final paint = %v, want the committed .25", got)
	}
	if r.Active() {
		t.Fatal("a cancelled transition still reported active")
	}
	for i := 0; i < 3; i++ {
		at = at.Add(200 * time.Millisecond)
		if damage := r.Step(d, at); !damage.Empty() {
			t.Fatalf("a cancelled transition damaged %v", damage.Bounds())
		}
		if got := d.Ops[0].R; !near(got, .25) {
			t.Fatalf("a cancelled transition rewrote the paint to %v", got)
		}
	}
}

// TestPropertyRemovalCancels checks that dropping the transition declaration
// leaves the committed value and releases the renderer.
func TestPropertyRemovalCancels(t *testing.T) {
	r, start, middle := startRunningTransition(t)
	next := transitionChip(.25, 1, 0)
	if captures := r.CaptureTransitions(middle, next); len(captures) != 0 {
		t.Fatalf("captures without a declaration = %d, want 0", len(captures))
	}
	r.BindTransitions(next, nil, start.Add(500*time.Millisecond))
	assertCancelled(t, r, next, start.Add(500*time.Millisecond))
}

// TestElementRemovalCancels checks that a display without the element releases
// its renderer.
func TestElementRemovalCancels(t *testing.T) {
	r, start, middle := startRunningTransition(t)
	next := &layout.Display{Width: testWidth, Height: testHeight,
		PointsPerPixel: 1, PixelPerPoint: 1}
	r.CaptureTransitions(middle, next)
	r.BindTransitions(next, nil, start.Add(500*time.Millisecond))
	if got := r.Stats().TransitionsCancelled; got != 1 {
		t.Fatalf("cancelled = %d, want 1", got)
	}
	if got := r.Stats().TransitionsBound; got != 0 {
		t.Fatalf("bound = %d, want 0", got)
	}
	if r.Active() {
		t.Fatal("a removed element still reported an active transition")
	}
	if _, ok := r.NextFrame(start.Add(500 * time.Millisecond)); ok {
		t.Fatal("a removed element asked for a frame")
	}
}

// TestOpReplacementCancels checks that a redraw which changes the operation
// kind releases the renderer and leaves the committed paint.
func TestOpReplacementCancels(t *testing.T) {
	r, start, middle := startRunningTransition(t)
	record := layout.Transition{Property: "background-color", DurationMS: 1000,
		TimingFunction: "linear"}
	next := transitionChip(.25, 1, 0, record)
	next.Ops[0].Kind = layout.DisplayOpText
	captures := r.CaptureTransitions(middle, next)
	r.BindTransitions(next, captures, start.Add(500*time.Millisecond))
	assertCancelled(t, r, next, start.Add(500*time.Millisecond))
}

// TestNewRunWithoutDeclarationCancels checks that a value change with no
// matching transition record cancels rather than binds.
func TestNewRunWithoutDeclarationCancels(t *testing.T) {
	r, start, middle := startRunningTransition(t)
	next := transitionChip(.25, 1, 0, linearTransition("opacity"))
	captures := r.CaptureTransitions(middle, next)
	r.BindTransitions(next, captures, start.Add(500*time.Millisecond))
	assertCancelled(t, r, next, start.Add(500*time.Millisecond))
}

// TestSnapshotToCurrentValueCancels checks the unchanged-paint rule: a new
// target that equals the presented paint but differs from the running
// destination cancels with no replacement, even though the two displays carry
// the same paint.
func TestSnapshotToCurrentValueCancels(t *testing.T) {
	r, start, middle := startRunningTransition(t)
	record := layout.Transition{Property: "background-color", DurationMS: 1000,
		TimingFunction: "linear"}
	next := transitionChip(.5, 1, 0, record)
	captures := r.CaptureTransitions(middle, next)
	if len(captures) != 1 || !captures[0].cancel {
		t.Fatalf("captures = %+v, want one cancel", captures)
	}
	r.BindTransitions(next, captures, start.Add(500*time.Millisecond))
	if got := r.Stats().TransitionsStarted; got != 0 {
		t.Fatalf("started = %d, want 0", got)
	}
	if got := next.Ops[0].R; !near(got, .5) {
		t.Fatalf("paint = %v, want the committed .5", got)
	}
	if damage := r.Step(next, start.Add(900*time.Millisecond)); !damage.Empty() {
		t.Fatalf("a cancelled run damaged %v", damage.Bounds())
	}
}

func TestEasingFunctions(t *testing.T) {
	for _, test := range []struct {
		name string
		at   float64
		want float64
	}{
		{name: "linear", at: .5, want: .5},
		{name: "steps(4, end)", at: .5, want: .5},
		{name: "linear(0, 1)", at: .25, want: .25},
	} {
		if got := Ease(test.name, test.at); got != test.want {
			t.Errorf("Ease(%q, %g) = %g, want %g", test.name, test.at, got, test.want)
		}
	}
	if got := Ease("ease", .5); got <= 0 || got >= 1 {
		t.Fatalf("ease midpoint = %g, want an interior value", got)
	}
}

// transitionInterruptEpsilon is the paint error a browser comparison accepts.
// The expected values below are the browser's own eased samples rounded to
// five decimals.
const transitionInterruptEpsilon = 2e-3

// startColorTransition binds one colour run from 0 to 1 and returns the
// runtime, the clock, and the two displays.
func startColorTransition(t *testing.T, record layout.Transition) (*Runtime, time.Time, *layout.Display, *layout.Display) {
	t.Helper()

	old := transitionChip(0, 1, 0, record)
	next := transitionChip(1, 1, 0, record)
	r := New()
	start := testClock()
	r.now = func() time.Time { return start }
	captures := r.CaptureTransitions(old, next)
	if len(captures) != 1 {
		t.Fatalf("captures = %d, want 1", len(captures))
	}
	r.BindTransitions(next, captures, start)

	return r, start, old, next
}

// TestInterruptShortensTheReverseByProgress interrupts a linear run at 25, 50,
// and 75 percent and checks the shortened reverse against the browser's
// timestamps: the reverse starts at the presented value, moves half-way at
// half of the shortened duration, and reaches the original from value at the
// shortened end, not at the declared end.
func TestInterruptShortensTheReverseByProgress(t *testing.T) {
	for _, test := range []struct {
		name     string
		fraction float64
	}{
		{name: "25 percent", fraction: .25},
		{name: "50 percent", fraction: .5},
		{name: "75 percent", fraction: .75},
	} {
		t.Run(test.name, func(t *testing.T) {
			record := layout.Transition{Property: "background-color", DurationMS: 1000,
				TimingFunction: "linear"}
			r, start, _, forward := startColorTransition(t, record)

			elapsed := time.Duration(test.fraction*1000) * time.Millisecond
			interrupt := start.Add(elapsed)
			r.Step(forward, interrupt)
			if got := forward.Ops[0].R; !near(got, test.fraction) {
				t.Fatalf("presented paint = %v, want %v", got, test.fraction)
			}

			back := transitionChip(0, 1, 0, record)
			captures := r.CaptureTransitions(forward, back)
			if len(captures) != 1 {
				t.Fatalf("interrupt captures = %d, want 1", len(captures))
			}
			r.BindTransitions(back, captures, interrupt)
			if got := r.Stats().TransitionsStarted; got != 1 {
				t.Fatalf("started = %d, want the reversed run", got)
			}
			if got := r.Stats().TransitionsCancelled; got != 1 {
				t.Fatalf("cancelled = %d, want the replaced run", got)
			}

			shortened := time.Duration(test.fraction*1000) * time.Millisecond
			r.Step(back, interrupt)
			if got := back.Ops[0].R; !near(got, test.fraction) {
				t.Fatalf("reverse start = %v, want the presented %v", got, test.fraction)
			}
			r.Step(back, interrupt.Add(shortened/2))
			if got := back.Ops[0].R; !near(got, test.fraction/2) {
				t.Fatalf("reverse midpoint = %v, want %v", got, test.fraction/2)
			}
			r.Step(back, interrupt.Add(shortened))
			if got := back.Ops[0].R; !near(got, 0) {
				t.Fatalf("reverse end = %v, want 0 at the shortened deadline", got)
			}
			if r.Active() {
				t.Fatal("a finished reverse still reported active")
			}
			if _, ok := r.NextFrame(interrupt.Add(shortened)); ok {
				t.Fatal("a finished reverse asked for another frame")
			}
		})
	}
}

// TestInterruptWithEaseMatchesBrowserNumbers repeats the Chromium
// reverse-transition-with-easing sequence at 100->200 width, normalized to an
// opacity-like 0->1 paint. The reverse shortening factor is the timing
// function's output (0.68254), not the raw time fraction, and the reverse
// keeps the declared easing.
func TestInterruptWithEaseMatchesBrowserNumbers(t *testing.T) {
	record := layout.Transition{Property: "background-color", DurationMS: 1000,
		TimingFunction: "ease"}
	r, start, _, forward := startColorTransition(t, record)

	forwardAt := start.Add(400 * time.Millisecond)
	r.Step(forward, forwardAt)
	if got := forward.Ops[0].R; math.Abs(got-.68254) > transitionInterruptEpsilon {
		t.Fatalf("forward sample = %v, want the browser's 0.68254", got)
	}

	back := transitionChip(0, 1, 0, record)
	captures := r.CaptureTransitions(forward, back)
	if len(captures) != 1 {
		t.Fatalf("first reverse captures = %d, want 1", len(captures))
	}
	r.BindTransitions(back, captures, forwardAt)

	firstReverseAt := start.Add(600 * time.Millisecond)
	r.Step(back, firstReverseAt)
	if got := back.Ops[0].R; math.Abs(got-.34165) > transitionInterruptEpsilon {
		t.Fatalf("first reverse sample = %v, want the browser's 0.34165", got)
	}

	again := transitionChip(1, 1, 0, record)
	captures = r.CaptureTransitions(back, again)
	if len(captures) != 1 {
		t.Fatalf("second reverse captures = %d, want 1", len(captures))
	}
	r.BindTransitions(again, captures, firstReverseAt)
	if got := r.Stats().TransitionsStarted; got != 1 {
		t.Fatalf("second reverse started = %d, want 1", got)
	}

	// The reversing-adjusted start value of the second run is the first
	// reverse's destination (1), so the browser shortens by 0.65831.
	secondReverseAt := start.Add(time.Second)
	r.Step(again, secondReverseAt)
	if got := again.Ops[0].R; math.Abs(got-.92779) > transitionInterruptEpsilon {
		t.Fatalf("second reverse sample = %v, want the browser's 0.92779", got)
	}
	r.Step(again, start.Add(1400*time.Millisecond))
	if got := again.Ops[0].R; !near(got, 1) {
		t.Fatalf("second reverse end = %v, want 1", got)
	}
}

// TestThirdValueChangeUsesTheFullDuration checks the non-reversal branch: a new
// destination that is neither the run's start nor its end starts from the
// presented paint with the whole declared duration, not a shortened one.
func TestThirdValueChangeUsesTheFullDuration(t *testing.T) {
	record := layout.Transition{Property: "background-color", DurationMS: 1000,
		TimingFunction: "linear"}
	r, start, _, forward := startColorTransition(t, record)

	interrupt := start.Add(500 * time.Millisecond)
	r.Step(forward, interrupt)
	if got := forward.Ops[0].R; !near(got, .5) {
		t.Fatalf("presented paint = %v, want .5", got)
	}

	third := transitionChip(.25, 1, 0, record)
	captures := r.CaptureTransitions(forward, third)
	if len(captures) != 1 {
		t.Fatalf("third-value captures = %d, want 1", len(captures))
	}
	r.BindTransitions(third, captures, interrupt)

	r.Step(third, interrupt.Add(500*time.Millisecond))
	if got := third.Ops[0].R; !near(got, .375) {
		t.Fatalf("third-value midpoint = %v, want .375 from the full duration", got)
	}
	r.Step(third, interrupt.Add(1000*time.Millisecond))
	if got := third.Ops[0].R; !near(got, .25) {
		t.Fatalf("third-value end = %v, want .25", got)
	}
}

// TestInterruptInsideTheDelayCancelsWithoutMoving checks the delay boundary: a
// run that has not started presents its from paint, so a change back to that
// value is a cancel rather than a reverse. The committed paint is already the
// answer and the pending destination must not fire later.
func TestInterruptInsideTheDelayCancelsWithoutMoving(t *testing.T) {
	record := layout.Transition{Property: "background-color", DurationMS: 1000,
		DelayMS: 200, TimingFunction: "linear"}
	old := transitionChip(0, 1, 0, record)
	forward := transitionChip(1, 1, 0, record)

	r := New()
	start := testClock()
	r.now = func() time.Time { return start }
	captures := r.CaptureTransitions(old, forward)
	r.BindTransitions(forward, captures, start)

	interrupt := start.Add(50 * time.Millisecond)
	r.Step(forward, interrupt)
	if got := forward.Ops[0].R; !near(got, 0) {
		t.Fatalf("delayed paint = %v, want the from value 0", got)
	}

	back := transitionChip(0, 1, 0, record)
	captures = r.CaptureTransitions(forward, back)
	if len(captures) != 1 || !captures[0].cancel {
		t.Fatalf("delay interrupt captures = %+v, want one cancel", captures)
	}
	r.BindTransitions(back, captures, interrupt)
	if got := r.Stats().TransitionsStarted; got != 0 {
		t.Fatalf("delay interrupt started = %d, want 0", got)
	}
	if got := r.Stats().TransitionsBound; got != 0 {
		t.Fatalf("delay interrupt bound = %d, want 0", got)
	}
	if got := r.Stats().TransitionsCancelled; got != 1 {
		t.Fatalf("delay interrupt cancelled = %d, want 1", got)
	}
	if r.Active() {
		t.Fatal("a cancelled delayed run still reported active")
	}
	if _, ok := r.NextFrame(interrupt); ok {
		t.Fatal("a cancelled delayed run asked for a frame")
	}

	// The old destination must never fire after the delay window.
	later := interrupt.Add(1000 * time.Millisecond)
	if damage := r.Step(back, later); !damage.Empty() {
		t.Fatalf("a cancelled delayed run damaged %v", damage.Bounds())
	}
	if got := back.Ops[0].R; !near(got, 0) {
		t.Fatalf("paint after the old deadline = %v, want 0", got)
	}
}

// TestZeroDurationRunSnapsAndCancels checks both zero-duration rules: a zero
// declared run installs no renderer because the commit already carries the
// destination, and it cancels the run it replaces.
func TestZeroDurationRunSnapsAndCancels(t *testing.T) {
	running := layout.Transition{Property: "background-color", DurationMS: 1000,
		TimingFunction: "linear"}
	r, start, _, forward := startColorTransition(t, running)
	interrupt := start.Add(500 * time.Millisecond)
	r.Step(forward, interrupt)
	if got := forward.Ops[0].R; !near(got, .5) {
		t.Fatalf("presented paint = %v, want .5", got)
	}

	instant := layout.Transition{Property: "background-color", DurationMS: 0,
		TimingFunction: "linear"}
	next := transitionChip(.25, 1, 0, instant)
	captures := r.CaptureTransitions(forward, next)
	if len(captures) != 1 {
		t.Fatalf("zero-duration captures = %d, want 1", len(captures))
	}
	r.BindTransitions(next, captures, interrupt)
	if got := r.Stats().TransitionsStarted; got != 0 {
		t.Fatalf("zero-duration started = %d, want 0", got)
	}
	if got := r.Stats().TransitionsBound; got != 0 {
		t.Fatalf("zero-duration bound = %d, want 0", got)
	}
	if got := r.Stats().TransitionsCancelled; got != 1 {
		t.Fatalf("zero-duration cancelled = %d, want 1", got)
	}
	if got := next.Ops[0].R; !near(got, .25) {
		t.Fatalf("zero-duration paint = %v, want the committed .25", got)
	}
	if r.Active() {
		t.Fatal("a zero-duration run still reported active")
	}
	if _, ok := r.NextFrame(interrupt); ok {
		t.Fatal("a zero-duration run asked for a frame")
	}

	if damage := r.Step(next, interrupt.Add(time.Second)); !damage.Empty() {
		t.Fatalf("a zero-duration run damaged %v", damage.Bounds())
	}
	if got := next.Ops[0].R; !near(got, .25) {
		t.Fatalf("zero-duration paint moved to %v, want .25", got)
	}
}

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

// fadeAnimation is one 100ms linear opacity animation from 0 to 1 with both
// fills, so it owns the field before the first sample and after it finishes.
func fadeAnimation() (layout.Animation, layout.KeyframesRule) {
	return layout.Animation{
			Name: "fade", DurationMS: transitionFrameMS, IterationCount: 1,
			TimingFunction: "linear", FillMode: "both", PlayState: "running",
		}, layout.KeyframesRule{
			Name: "fade",
			Keyframes: []layout.Keyframe{
				{Offset: 0, Declarations: []layout.KeyframeDeclaration{
					{Property: "opacity", Value: "0"}}},
				{Offset: 1, Declarations: []layout.KeyframeDeclaration{
					{Property: "opacity", Value: "1"}}},
			},
		}
}

// TestAnimationPrecedenceSkipsTheTransitionCapture checks the cascade rule: an
// animation on opacity owns the field, so the transition capture is counted
// and never becomes a renderer. The animation's value is what every sampled
// frame shows, including after it finishes.
func TestAnimationPrecedenceSkipsTheTransitionCapture(t *testing.T) {
	record := linearTransition("opacity")
	old := transitionChip(1, 1, 0, record)
	next := transitionChip(1, .5, 0, record)
	animation, keyframes := fadeAnimation()
	next.Boxes[0].Animations = []layout.Animation{animation}
	next.Keyframes = []layout.KeyframesRule{keyframes}

	r := New()
	start := testClock()
	r.now = func() time.Time { return start }
	if err := r.Bind(next); err != nil {
		t.Fatalf("bind animation: %v", err)
	}
	captures := r.CaptureTransitions(old, next)
	if len(captures) != 1 {
		t.Fatalf("captures = %d, want the opacity change", len(captures))
	}
	r.BindTransitions(next, captures, start)
	if got := r.Stats().TransitionsStarted; got != 0 {
		t.Fatalf("started = %d, want the animation to win", got)
	}
	if got := r.Stats().TransitionsBound; got != 0 {
		t.Fatalf("bound = %d, want 0", got)
	}
	if got := r.Stats().TransitionsSkipped; got != 1 {
		t.Fatalf("skipped = %d, want the precedence skip counted", got)
	}

	half := start.Add(transitionFrameMS / 2 * time.Millisecond)
	r.Step(next, half)
	if got := next.Ops[0].Opacity(); !near(got, .5) {
		t.Fatalf("opacity at the midpoint = %v, want the animation's .5", got)
	}
	r.Step(next, start.Add(transitionFrameMS*time.Millisecond))
	if got := next.Ops[0].Opacity(); !near(got, 1) {
		t.Fatalf("opacity at the end = %v, want the animation's 1", got)
	}

	// A later commit cannot revive the skipped transition.
	r.BindTransitions(next, nil, start.Add(transitionFrameMS*time.Millisecond))
	if got := r.Stats().TransitionsStarted; got != 0 {
		t.Fatalf("a later commit started %d transitions, want 0", got)
	}
	r.Step(next, start.Add(transitionFrameMS*2*time.Millisecond))
	if got := next.Ops[0].Opacity(); !near(got, 1) {
		t.Fatalf("opacity after the animation = %v, want the fill's 1", got)
	}
}

// TestAnimationPrecedenceCancelsARunningTransition checks the other direction:
// when an animation appears while a transition is in flight, the transition is
// dropped and does not resume after the animation ends.
func TestAnimationPrecedenceCancelsARunningTransition(t *testing.T) {
	record := layout.Transition{Property: "opacity", DurationMS: 1000,
		TimingFunction: "linear"}
	old := transitionChip(1, 1, 0, record)
	next := transitionChip(1, .5, 0, record)

	r := New()
	start := testClock()
	r.now = func() time.Time { return start }
	captures := r.CaptureTransitions(old, next)
	r.BindTransitions(next, captures, start)
	interrupt := start.Add(250 * time.Millisecond)
	r.Step(next, interrupt)
	if got := next.Ops[0].Opacity(); !near(got, .875) {
		t.Fatalf("transition sample = %v, want .875", got)
	}

	animated := transitionChip(1, .5, 0, record)
	animation, keyframes := fadeAnimation()
	animated.Boxes[0].Animations = []layout.Animation{animation}
	animated.Keyframes = []layout.KeyframesRule{keyframes}
	r.now = func() time.Time { return interrupt }
	if err := r.Bind(animated); err != nil {
		t.Fatalf("bind animation: %v", err)
	}
	r.BindTransitions(animated, nil, interrupt)
	if got := r.Stats().TransitionsCancelled; got != 1 {
		t.Fatalf("cancelled = %d, want the transition dropped", got)
	}
	if got := r.Stats().TransitionsBound; got != 0 {
		t.Fatalf("bound = %d, want 0", got)
	}

	r.Step(animated, interrupt.Add(transitionFrameMS/2*time.Millisecond))
	if got := animated.Ops[0].Opacity(); !near(got, .5) {
		t.Fatalf("opacity at the animation midpoint = %v, want .5", got)
	}
	r.Step(animated, interrupt.Add(transitionFrameMS*time.Millisecond))
	if got := animated.Ops[0].Opacity(); !near(got, 1) {
		t.Fatalf("opacity after the animation = %v, want the fill's 1", got)
	}
	if got := r.Stats().TransitionsStarted; got != 0 {
		t.Fatalf("the cancelled transition revived: started = %d", got)
	}
}

// TestProgressForMatchesTheRendererClock pins the record-based accessor to the
// same delay, duration, and easing the renderer uses.
func TestProgressForMatchesTheRendererClock(t *testing.T) {
	start := time.Unix(20, 0)
	record := layout.Transition{DurationMS: 1000, DelayMS: 200, TimingFunction: "linear"}

	if got, active := ProgressFor(record, start, start.Add(100*time.Millisecond)); got != 0 || !active {
		t.Fatalf("progress inside the delay = %g/%v, want 0/true", got, active)
	}
	if got, active := ProgressFor(record, start, start.Add(700*time.Millisecond)); got != .5 || !active {
		t.Fatalf("progress at the midpoint = %g/%v, want .5/true", got, active)
	}
	if got, active := ProgressFor(record, start, start.Add(1200*time.Millisecond)); got != 1 || active {
		t.Fatalf("progress after the end = %g/%v, want 1/false", got, active)
	}

	eased := layout.Transition{DurationMS: 1000, TimingFunction: "ease"}
	if got, _ := ProgressFor(eased, start, start.Add(500*time.Millisecond)); got != Ease("ease", .5) {
		t.Fatalf("eased progress = %g, want %g", got, Ease("ease", .5))
	}

	instant := layout.Transition{}
	if got, active := ProgressFor(instant, start, start); got != 1 || active {
		t.Fatalf("zero-duration progress = %g/%v, want 1/false", got, active)
	}
}

// TestProgressUsesTheStepClock pins the renderer accessor to the same delay,
// duration, and easing Step uses.
func TestProgressUsesTheStepClock(t *testing.T) {
	start := time.Unix(10, 0)
	op := layout.DisplayOp{}

	linear := renderer(t, &op, time.Second, 200*time.Millisecond, "linear")
	if got := linear.Progress(start); got != 0 {
		t.Fatalf("progress before Start = %g, want 0", got)
	}
	linear.Start(start)
	if got := linear.Progress(start.Add(100 * time.Millisecond)); got != 0 {
		t.Fatalf("progress inside the delay = %g, want 0", got)
	}
	if got := linear.Progress(start.Add(700 * time.Millisecond)); got != .5 {
		t.Fatalf("progress at the midpoint = %g, want .5", got)
	}
	if got := linear.Progress(start.Add(1200 * time.Millisecond)); got != 1 {
		t.Fatalf("progress after the end = %g, want 1", got)
	}

	eased := renderer(t, &op, time.Second, 0, "ease")
	eased.Start(start)
	if got, want := eased.Progress(start.Add(500*time.Millisecond)),
		Ease("ease", .5); got != want {
		t.Fatalf("eased progress = %g, want the timing-function output %g", got, want)
	}

	negative := renderer(t, &op, time.Second, -500*time.Millisecond, "linear")
	negative.Start(start)
	if got := negative.Progress(start); got != .5 {
		t.Fatalf("negative-delay progress = %g, want .5", got)
	}

	instant := renderer(t, &op, 0, 0, "linear")
	instant.Start(start)
	if got := instant.Progress(start); got != 1 {
		t.Fatalf("zero-duration progress = %g, want 1", got)
	}
}

func TestStepAppliesDelayThenInterpolates(t *testing.T) {
	op := layout.DisplayOp{}
	r := renderer(t, &op, time.Second, 200*time.Millisecond, "linear")
	start := time.Unix(1, 0)
	r.Start(start)
	if !r.Step(start.Add(100*time.Millisecond)) || op.X != 0 {
		t.Fatal("positive delay did not hold the starting paint")
	}
	if !r.Step(start.Add(700*time.Millisecond)) || op.X != 50 {
		t.Fatalf("midpoint X = %g, want 50", op.X)
	}
	if r.Step(start.Add(1200*time.Millisecond)) || op.X != 100 {
		t.Fatalf("finished X = %g, want 100", op.X)
	}
}

func TestStepUsesNegativeDelayAndEasing(t *testing.T) {
	op := layout.DisplayOp{}
	r := renderer(t, &op, time.Second, -500*time.Millisecond, "linear")
	start := time.Unix(2, 0)
	r.Start(start)
	if !r.Step(start) || op.X != 50 {
		t.Fatalf("negative-delay X = %g, want 50", op.X)
	}
	r = renderer(t, &op, time.Second, 0, "steps(2, end)")
	r.Start(start)
	if !r.Step(start.Add(250*time.Millisecond)) || op.X != 0 {
		t.Fatalf("step easing X = %g, want 0", op.X)
	}
}

func TestZeroDurationWritesFinalPaint(t *testing.T) {
	op := layout.DisplayOp{}
	r := renderer(t, &op, 0, 0, "linear")
	r.Start(time.Unix(3, 0))
	if r.Step(time.Unix(3, 0)) || op.X != 100 {
		t.Fatalf("zero-duration X = %g, want 100", op.X)
	}
}

// TestRetargetMovesTracksWithoutRestarting keeps the clock across a display
// commit: the renderer writes the new operation at the same phase.
func TestRetargetMovesTracksWithoutRestarting(t *testing.T) {
	first := layout.DisplayOp{}
	next := layout.DisplayOp{}
	r, err := newTransitionRenderer([]Track{{Op: &first, From: Paint{Opacity: 1},
		To: Paint{X: 100, Opacity: 1}, Duration: time.Second, Easing: "linear"}})
	if err != nil {
		t.Fatal(err)
	}
	start := time.Unix(4, 0)
	r.Start(start)
	if !r.Step(start.Add(500*time.Millisecond)) || first.X != 50 {
		t.Fatalf("first sample X = %g, want 50", first.X)
	}
	if err := r.Retarget([]*layout.DisplayOp{&next}); err != nil {
		t.Fatal(err)
	}
	if !r.Step(start.Add(750*time.Millisecond)) || next.X != 75 {
		t.Fatalf("retargeted sample X = %g, want 75", next.X)
	}
}

func renderer(t *testing.T, op *layout.DisplayOp, duration, delay time.Duration, ease string) *Renderer {
	t.Helper()
	r, err := newTransitionRenderer([]Track{{Op: op, From: Paint{Opacity: 1},
		To: Paint{X: 100, Opacity: 0}, Duration: duration, Delay: delay, Easing: ease}})
	if err != nil {
		t.Fatal(err)
	}
	return r
}

// TestCaptureIgnoresStaleSharedIndex pins a crash seen in the Excel example:
// the runtime's cached index described a larger display than the previous
// paint, and indexFor repointed it because r.display matched. Selecting from
// it returned operation positions past the end of the previous display and
// sameOpKinds panicked with an index out of range.
func TestCaptureIgnoresStaleSharedIndex(t *testing.T) {
	records := []layout.Transition{linearTransition("opacity")}
	old := transitionChip(0, 1, 0, records...)
	next := transitionChip(0, .5, 0, records...)

	// A bigger earlier display leaves the cached index with extra operations
	// owned by the chip's box.
	big := transitionChip(0, 1, 0, records...)
	for range 3 {
		op := big.Ops[0]
		big.Ops = append(big.Ops, op)
		big.OpOwners = append(big.OpOwners, 0)
		big.Boxes[0].OperationIDs = append(big.Boxes[0].OperationIDs, op.ID)
	}

	r := New()
	r.index = NewIndex(big)
	r.display = old

	captures := r.CaptureTransitions(old, next)
	if len(captures) != 1 {
		t.Fatalf("captures = %d, want 1", len(captures))
	}
}
