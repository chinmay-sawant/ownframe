package transition

import (
	"testing"
	"time"

	"github.com/chinmay-sawant/blinkless/layout"
	"github.com/chinmay-sawant/ownframe/internal/cssmotion"
)

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
	if got, _ := ProgressFor(eased, start, start.Add(500*time.Millisecond)); got != cssmotion.Ease("ease", .5) {
		t.Fatalf("eased progress = %g, want %g", got, cssmotion.Ease("ease", .5))
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
		cssmotion.Ease("ease", .5); got != want {
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
