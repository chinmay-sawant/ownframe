package transition

import (
	"testing"
	"time"

	"github.com/chinmay-sawant/ownframe"
	"github.com/chinmay-sawant/ownframe/internal/cssmotion"
)

func TestStepAppliesDelayThenInterpolates(t *testing.T) {
	op := ownframe.DisplayOp{}
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
	op := ownframe.DisplayOp{}
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
	op := ownframe.DisplayOp{}
	r := renderer(t, &op, 0, 0, "linear")
	r.Start(time.Unix(3, 0))
	if r.Step(time.Unix(3, 0)) || op.X != 100 {
		t.Fatalf("zero-duration X = %g, want 100", op.X)
	}
}

func renderer(t *testing.T, op *ownframe.DisplayOp, duration, delay time.Duration, ease string) *Renderer {
	t.Helper()
	r, err := New([]Track{{Op: op, From: cssmotion.Paint{Opacity: 1},
		To: cssmotion.Paint{X: 100, Opacity: 0}, Duration: duration, Delay: delay, Easing: ease}})
	if err != nil {
		t.Fatal(err)
	}
	return r
}
