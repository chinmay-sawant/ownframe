package motion

import (
	"testing"
	"time"

	"github.com/chinmay-sawant/blinkless/layout"
)

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
