package frame

import (
	"math"
	"testing"
	"time"

	"github.com/chinmay-sawant/ownframe/internal/page"
)

func TestFadeOutStopsInvisible(t *testing.T) {
	t.Parallel()

	d, op := cardDisplay()
	a := FadeTo(d, []*page.DisplayOp{op}, 0, time.Millisecond, Linear)

	a.Step(t0)
	a.Step(t0.Add(time.Millisecond))

	if math.Abs(op.Alpha-minAlpha) > 1e-9 {
		t.Fatalf("alpha = %v", op.Alpha)
	}

	if op.Opacity() > 0.01 {
		t.Fatalf("opacity = %v", op.Opacity())
	}
}

func TestSetAlphaClamps(t *testing.T) {
	t.Parallel()

	_, op := cardDisplay()
	SetAlpha([]*page.DisplayOp{op}, 0)

	if op.Alpha != minAlpha {
		t.Fatalf("hidden alpha = %v", op.Alpha)
	}

	SetAlpha([]*page.DisplayOp{op}, 2)

	if op.Alpha != 1 {
		t.Fatalf("capped alpha = %v", op.Alpha)
	}
}
