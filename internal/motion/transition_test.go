package motion

import (
	"math"
	"testing"
	"time"

	"github.com/chinmay-sawant/blinkless/layout"
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
// rather than the next vsync.
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
