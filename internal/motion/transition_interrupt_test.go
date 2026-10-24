package motion

import (
	"math"
	"testing"
	"time"

	"github.com/chinmay-sawant/blinkless/layout"
)

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
