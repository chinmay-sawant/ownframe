package frame

import (
	"math"
	"testing"
	"time"

	"github.com/chinmay-sawant/ownframe"
)

func TestInterruptedTranslateRetargets(t *testing.T) {
	t.Parallel()

	d, op := cardDisplay()
	a := TranslateTo(d, []*ownframe.DisplayOp{op}, 30, 0, 100*time.Millisecond, Linear)
	a.Step(t0)
	a.Step(t0.Add(50 * time.Millisecond))

	if op.X != 25 {
		t.Fatalf("mid x = %v", op.X)
	}

	b := TranslateTo(d, []*ownframe.DisplayOp{op}, -10, 0, 100*time.Millisecond, Linear)
	b.Step(t0.Add(50 * time.Millisecond))

	if op.X != 25 {
		t.Fatalf("retarget jumped to %v", op.X)
	}

	b.Step(t0.Add(100 * time.Millisecond))

	if op.X != 20 {
		t.Fatalf("retargeted x = %v", op.X)
	}

	b.Step(t0.Add(150 * time.Millisecond))

	if op.X != 15 {
		t.Fatalf("final x = %v", op.X)
	}
}

func TestInterruptedFadeRetargets(t *testing.T) {
	t.Parallel()

	d, op := cardDisplay()
	a := FadeTo(d, []*ownframe.DisplayOp{op}, 0.2, 100*time.Millisecond, Linear)
	a.Step(t0)
	a.Step(t0.Add(50 * time.Millisecond))

	b := FadeTo(d, []*ownframe.DisplayOp{op}, 0, 100*time.Millisecond, Linear)
	b.Step(t0.Add(50 * time.Millisecond))

	if math.Abs(op.Alpha-0.6) > 1e-9 {
		t.Fatalf("retarget jumped to %v", op.Alpha)
	}
}
