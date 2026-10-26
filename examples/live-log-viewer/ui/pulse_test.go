package ui

import (
	"context"
	"testing"
	"time"
)

func TestPulseAnimatesRetainedOp(t *testing.T) {
	a := newHeadless(t)
	a.follow.SetFollow(true)
	ctx := context.Background()

	if err := a.draw(ctx); err != nil {
		t.Fatal(err)
	}

	a.pulse(time.Now())

	if a.pulseDot == nil {
		t.Fatal("live dot operation not found")
	}

	if alpha := a.pulseDot.Alpha; alpha <= 0.3 || alpha > 0.8 {
		t.Fatalf("pulse alpha = %v", alpha)
	}

	gen := a.pulseGen

	// Resize so the redraw replaces the display list. An identical redraw
	// keeps the list and the binding, so there would be nothing to rebind.
	w, h := a.page.Size()
	a.page.SetSize(w+10, h+10)
	if err := a.draw(ctx); err != nil {
		t.Fatal(err)
	}

	a.pulse(time.Now())

	if a.pulseGen == gen {
		t.Fatal("pulse generation did not advance after a redraw")
	}

	if a.pulseDot == nil {
		t.Fatal("live dot operation not rebound after a redraw")
	}
}
