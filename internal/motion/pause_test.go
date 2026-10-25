package motion

import (
	"testing"
	"time"

	"github.com/chinmay-sawant/blinkless/layout"
)

// splitDisplay is one card with a transform animation on its subtree and a
// background-colour animation on its own fill. The two properties bind as two
// entries that share the fill operation, which is the case the pause handoff
// has to split by paint mask.
func splitDisplay(playState string) *layout.Display {
	return &layout.Display{
		Width: testWidth, Height: testHeight, PointsPerPixel: 1, PixelPerPoint: 1,
		Ops: []layout.DisplayOp{
			{ID: 1, Kind: layout.DisplayOpFillRect,
				X: 10, Y: 10, W: 20, H: 20, R: 0.1, G: 0.2, B: 0.3, Alpha: 1},
			{ID: 2, Kind: layout.DisplayOpText,
				X: 12, Y: 12, W: 10, H: 10, R: 1, G: 1, B: 1, Alpha: 1},
		},
		OpOwners: []int{0, 0},
		Boxes: []layout.Box{{
			ID: "card", Tag: "div", X: 10, Y: 10, W: 20, H: 20, ParentIndex: -1,
			OperationIDs: []uint64{1, 2},
			Animations: []layout.Animation{
				{Name: "slide", DurationMS: testDurationMS, IterationInfinite: true,
					PlayState: playState, FillMode: "both"},
				{Name: "tint", DurationMS: testDurationMS, IterationInfinite: true,
					PlayState: playState, FillMode: "both"},
			},
		}},
		Keyframes: []layout.KeyframesRule{
			{Name: "slide", Keyframes: []layout.Keyframe{
				{Offset: 0, Declarations: []layout.KeyframeDeclaration{
					{Property: "transform", Value: "translateX(0px)"}}},
				{Offset: 1, Declarations: []layout.KeyframeDeclaration{
					{Property: "transform", Value: "translateX(20px)"}}},
			}},
			{Name: "tint", Keyframes: []layout.Keyframe{
				{Offset: 0, Declarations: []layout.KeyframeDeclaration{
					{Property: "background-color", Value: "rgb(0,0,0)"}}},
				{Offset: 1, Declarations: []layout.KeyframeDeclaration{
					{Property: "background-color", Value: "rgb(255,255,255)"}}},
			}},
		},
	}
}

// TestPausedRebindKeepsTheDisplayedPose is the page bug in miniature: a redraw
// while paused publishes a display with the declared paint, and the new player
// must not snap the chip back to it before the next tick.
func TestPausedRebindKeepsTheDisplayedPose(t *testing.T) {
	r := New()
	r.now = testClock
	first := animatedDisplay("running", true)

	if err := r.Bind(first); err != nil {
		t.Fatalf("bind: %v", err)
	}
	r.Step(first, testClock().Add(40*time.Millisecond))
	if got, want := movedBy(&first.Ops[0]), float64(8); got != want {
		t.Fatalf("sample before pause = %v, want %v", got, want)
	}

	paused := animatedDisplay("paused", true)
	r.now = func() time.Time { return testClock().Add(80 * time.Millisecond) }
	if err := r.BindAfter(first, paused); err != nil {
		t.Fatalf("rebind: %v", err)
	}

	// The redraw happens between ticks, so the pose must already be copied.
	if got, want := movedBy(&paused.Ops[0]), float64(8); got != want {
		t.Fatalf("pose immediately after the paused publish = %v, want %v", got, want)
	}

	r.Step(paused, testClock().Add(500*time.Millisecond))
	if got, want := movedBy(&paused.Ops[0]), float64(8); got != want {
		t.Fatalf("paused sample = %v, want the frozen %v", got, want)
	}
	if r.Active() || r.Stats().Active != 0 {
		t.Fatalf("a paused runtime reports active=%v count=%d", r.Active(), r.Stats().Active)
	}

	if damage := r.Step(paused, testClock().Add(600*time.Millisecond)); !damage.Empty() {
		t.Fatalf("a second paused tick dirtied %v", damage.Bounds())
	}
}

// TestPausedPoseSurvivesTwoPublishesBeforeATick covers a redraw that follows
// another redraw while the animation is still paused: the second published
// display is built from the declared paint, but the frozen pose must come from
// the first paused display rather than the base.
func TestPausedPoseSurvivesTwoPublishesBeforeATick(t *testing.T) {
	r := New()
	r.now = testClock
	first := animatedDisplay("running", true)

	if err := r.Bind(first); err != nil {
		t.Fatalf("bind: %v", err)
	}
	r.Step(first, testClock().Add(40*time.Millisecond))

	paused := animatedDisplay("paused", true)
	r.now = func() time.Time { return testClock().Add(80 * time.Millisecond) }
	if err := r.BindAfter(first, paused); err != nil {
		t.Fatalf("first pause rebind: %v", err)
	}

	// No Step between the two publishes, so the first paused entry was never
	// sampled; its operations still carry the copied pose.
	again := animatedDisplay("paused", true)
	if err := r.BindAfter(paused, again); err != nil {
		t.Fatalf("second pause rebind: %v", err)
	}
	if got, want := movedBy(&again.Ops[0]), float64(8); got != want {
		t.Fatalf("pose after the second paused publish = %v, want %v", got, want)
	}
}

// TestResumeContinuesFromThePausedPhase resumes 500 ms after the pause and
// compares the sample with an uninterrupted run at the same active time.
func TestResumeContinuesFromThePausedPhase(t *testing.T) {
	r := New()
	r.now = testClock
	first := animatedDisplay("running", true)

	if err := r.Bind(first); err != nil {
		t.Fatalf("bind: %v", err)
	}
	r.Step(first, testClock().Add(40*time.Millisecond))

	paused := animatedDisplay("paused", true)
	r.now = func() time.Time { return testClock().Add(80 * time.Millisecond) }
	if err := r.BindAfter(first, paused); err != nil {
		t.Fatalf("pause rebind: %v", err)
	}
	r.Step(paused, testClock().Add(540*time.Millisecond))

	running := animatedDisplay("running", true)
	r.now = func() time.Time { return testClock().Add(540 * time.Millisecond) }
	if err := r.BindAfter(paused, running); err != nil {
		t.Fatalf("resume rebind: %v", err)
	}
	r.Step(running, testClock().Add(560*time.Millisecond))
	if got, want := movedBy(&running.Ops[0]), float64(12); got != want {
		t.Fatalf("resumed sample = %v, want %v after 60 ms of active time", got, want)
	}

	uninterrupted := New()
	uninterrupted.now = testClock
	fresh := animatedDisplay("running", true)
	if err := uninterrupted.Bind(fresh); err != nil {
		t.Fatalf("bind uninterrupted: %v", err)
	}
	uninterrupted.Step(fresh, testClock().Add(60*time.Millisecond))
	if got, want := movedBy(&fresh.Ops[0]), movedBy(&running.Ops[0]); got != want {
		t.Fatalf("uninterrupted active-time sample = %v, resumed = %v", got, want)
	}
}

// TestPausedPoseCopiesOnlyTheFieldsEachPlayerOwns checks the split handoff: the
// transform entry carries the matrix and the colour entry carries the fill
// channels, and neither overwrites the other or the label text.
func TestPausedPoseCopiesOnlyTheFieldsEachPlayerOwns(t *testing.T) {
	r := New()
	r.now = testClock
	first := splitDisplay("running")

	if err := r.Bind(first); err != nil {
		t.Fatalf("bind: %v", err)
	}
	r.Step(first, testClock().Add(40*time.Millisecond))
	if got, want := movedBy(&first.Ops[0]), float64(8); got != want {
		t.Fatalf("fill translation = %v, want %v", got, want)
	}
	if got, want := first.Ops[0].R, 0.4; got != want {
		t.Fatalf("fill red = %v, want %v", got, want)
	}

	paused := splitDisplay("paused")
	r.now = func() time.Time { return testClock().Add(80 * time.Millisecond) }
	if err := r.BindAfter(first, paused); err != nil {
		t.Fatalf("rebind: %v", err)
	}

	if got, want := movedBy(&paused.Ops[0]), float64(8); got != want {
		t.Fatalf("paused fill translation = %v, want %v", got, want)
	}
	if got, want := paused.Ops[0].R, 0.4; got != want {
		t.Fatalf("paused fill red = %v, want %v", got, want)
	}
	if got, want := movedBy(&paused.Ops[1]), float64(8); got != want {
		t.Fatalf("paused label translation = %v, want %v", got, want)
	}
	if paused.Ops[1].R != 1 || paused.Ops[1].G != 1 || paused.Ops[1].B != 1 {
		t.Fatalf("paused label colour = %v,%v,%v, want its own 1,1,1",
			paused.Ops[1].R, paused.Ops[1].G, paused.Ops[1].B)
	}

	r.Step(paused, testClock().Add(500*time.Millisecond))
	if got, want := movedBy(&paused.Ops[0]), float64(8); got != want {
		t.Fatalf("paused sampled translation = %v, want %v", got, want)
	}
	if got, want := paused.Ops[0].R, 0.4; got != want {
		t.Fatalf("paused sampled red = %v, want %v", got, want)
	}
}

// delayDisplay is animatedDisplay with the cycle and signed start delay the
// pause drift needs.
func delayDisplay(playState string, durationMS, delayMS float64) *layout.Display {
	d := animatedDisplay(playState, true)
	d.Boxes[0].Animations[0].DurationMS = durationMS
	d.Boxes[0].Animations[0].DelayStartMS = delayMS

	return d
}

// sampledDisplay is the display a re-projection patch hands back. The engine
// rebuilds it from the retained layout result, and the previous sample wrote
// through that result's operations, so the list still carries the sampled
// paint instead of the declared one. The copy is what makes the test model
// that handoff: writing through the next display must not move the previous
// one.
func sampledDisplay(d *layout.Display) *layout.Display {
	if d == nil {
		return nil
	}
	clone := *d
	clone.Ops = append([]layout.DisplayOp(nil), d.Ops...)
	clone.OpOwners = append([]int(nil), d.OpOwners...)

	return &clone
}

// TestPausedNegativeDelayHoldsAcrossSampledRebinds is the drift case: a paused
// animation whose negative delay already advanced it a quarter of the cycle is
// rebound five times onto displays that still carry its sample. The frozen
// pose must not move.
func TestPausedNegativeDelayHoldsAcrossSampledRebinds(t *testing.T) {
	r := New()
	r.now = testClock
	first := delayDisplay("paused", 4000, -1000)

	if err := r.Bind(first); err != nil {
		t.Fatalf("bind: %v", err)
	}
	r.Step(first, testClock().Add(10*time.Millisecond))
	if got, want := movedBy(&first.Ops[0]), float64(5); got != want {
		t.Fatalf("first paused sample = %v, want %v at a quarter of the shift", got, want)
	}

	previous := first
	for i := 0; i < 5; i++ {
		next := sampledDisplay(previous)
		at := testClock().Add(time.Duration(10+i*10) * time.Millisecond)
		r.now = func() time.Time { return at }
		if err := r.BindAfter(previous, next); err != nil {
			t.Fatalf("rebind %d: %v", i, err)
		}
		r.Step(next, at.Add(time.Millisecond))
		if got, want := movedBy(&next.Ops[0]), float64(5); got != want {
			t.Fatalf("rebind %d pose = %v, want the frozen %v", i, got, want)
		}
		previous = next
	}
}

// TestPausedPositiveDelayHoldsAcrossSampledRebinds keeps the other sign
// honest: while the animation is still inside its delay the frozen pose is the
// backwards fill, and a rebind must not advance it into the active interval.
func TestPausedPositiveDelayHoldsAcrossSampledRebinds(t *testing.T) {
	r := New()
	r.now = testClock
	first := delayDisplay("paused", 4000, 1000)

	if err := r.Bind(first); err != nil {
		t.Fatalf("bind: %v", err)
	}
	r.Step(first, testClock().Add(10*time.Millisecond))
	if got := movedBy(&first.Ops[0]); got != 0 {
		t.Fatalf("paused delay pose = %v, want the backwards fill 0", got)
	}

	previous := first
	for i := 0; i < 5; i++ {
		next := sampledDisplay(previous)
		at := testClock().Add(time.Duration(10+i*10) * time.Millisecond)
		r.now = func() time.Time { return at }
		if err := r.BindAfter(previous, next); err != nil {
			t.Fatalf("rebind %d: %v", i, err)
		}
		r.Step(next, at.Add(time.Millisecond))
		if got := movedBy(&next.Ops[0]); got != 0 {
			t.Fatalf("rebind %d pose = %v, want the delayed 0", i, got)
		}
		previous = next
	}
}

// TestRunningRebindOnSampledDisplayKeepsPhase is the same handoff without a
// pause: a running player bound to a display that still carries its last
// sample must advance by the elapsed time, not by the sample again.
func TestRunningRebindOnSampledDisplayKeepsPhase(t *testing.T) {
	r := New()
	r.now = testClock
	first := animatedDisplay("running", true)

	if err := r.Bind(first); err != nil {
		t.Fatalf("bind: %v", err)
	}
	r.Step(first, testClock().Add(20*time.Millisecond))
	if got, want := movedBy(&first.Ops[0]), float64(4); got != want {
		t.Fatalf("first sample = %v, want %v", got, want)
	}

	previous := first
	for i, want := range []float64{6, 8, 10, 12} {
		next := sampledDisplay(previous)
		at := testClock().Add(time.Duration(30+10*i) * time.Millisecond)
		r.now = func() time.Time { return at }
		if err := r.BindAfter(previous, next); err != nil {
			t.Fatalf("rebind %d: %v", i, err)
		}
		r.Step(next, at)
		if got := movedBy(&next.Ops[0]); got != want {
			t.Fatalf("rebind %d pose = %v, want %v at the same active time", i, got, want)
		}
		previous = next
	}
}

// restampDisplay is a page with two paused animations: a 3D rotateY card that
// asks for the engine's re-projection pass every tick, and a 2D chip with a
// negative delay whose pose must not drift when that pass rebinds it.
func restampDisplay() *layout.Display {
	d := delayDisplay("paused", 4000, -1000)
	d.Boxes = append(d.Boxes, layout.Box{
		ID: "flip", Tag: "div", X: 50, Y: 50, W: 20, H: 20, ParentIndex: -1,
		OperationIDs: []uint64{testOpID + 1},
		Animations: []layout.Animation{{
			Name: "flip", DurationMS: 4000, DelayStartMS: -1000,
			IterationInfinite: true, FillMode: "both", PlayState: "paused",
		}},
	})
	d.Ops = append(d.Ops, layout.DisplayOp{
		ID: testOpID + 1, Kind: layout.DisplayOpFillRect,
		X: 50, Y: 50, W: 20, H: 20, R: 1,
	})
	d.OpOwners = append(d.OpOwners, 1)
	d.Keyframes = append(d.Keyframes, layout.KeyframesRule{
		Name: "flip",
		Keyframes: []layout.Keyframe{
			{Offset: 0, Declarations: []layout.KeyframeDeclaration{
				{Property: testTransform, Value: "rotateY(0deg)"}}},
			{Offset: 1, Declarations: []layout.KeyframeDeclaration{
				{Property: testTransform, Value: "rotateY(180deg)"}}},
		},
	})

	return d
}

// TestPaused3DRestampDoesNotDriftThePausedPose rebinds the whole page on every
// tick, as the page's re-projection pass does. The paused 2D chip keeps its
// pose, and the 3D card keeps asking for the same re-projection without writing
// a stale two-dimensional matrix.
func TestPaused3DRestampDoesNotDriftThePausedPose(t *testing.T) {
	r := New()
	r.now = testClock
	first := restampDisplay()

	if err := r.Bind(first); err != nil {
		t.Fatalf("bind: %v", err)
	}
	r.Step(first, testClock().Add(10*time.Millisecond))
	if got, want := movedBy(&first.Ops[0]), float64(5); got != want {
		t.Fatalf("first chip pose = %v, want %v", got, want)
	}
	requests := r.TakeRestamp()
	if len(requests) != 1 || requests[0].ElementID != "flip" || requests[0].Transform == "" {
		t.Fatalf("first restamp requests = %+v, want the flip card's transform", requests)
	}
	want := requests[0].Transform

	previous := first
	for i := 0; i < 5; i++ {
		next := sampledDisplay(previous)
		at := testClock().Add(time.Duration(10+i*10) * time.Millisecond)
		r.now = func() time.Time { return at }
		if err := r.BindAfter(previous, next); err != nil {
			t.Fatalf("rebind %d: %v", i, err)
		}
		r.Step(next, at.Add(time.Millisecond))

		if got := movedBy(&next.Ops[0]); got != 5 {
			t.Fatalf("rebind %d chip pose = %v, want the frozen 5", i, got)
		}
		if next.Ops[1].XformSet {
			t.Fatalf("rebind %d wrote a stale 2D matrix for the 3D card: %+v",
				i, next.Ops[1].Transform())
		}
		requests = r.TakeRestamp()
		if len(requests) != 1 {
			t.Fatalf("rebind %d restamp requests = %+v, want one", i, requests)
		}
		if requests[0].Transform != want {
			t.Fatalf("rebind %d restamp transform = %q, want the frozen %q",
				i, requests[0].Transform, want)
		}
		previous = next
	}
}
