package main

import (
	"context"
	"math"
	"testing"
)

func TestClampBodyOffset(t *testing.T) {
	if got := clampBodyOffset(100, 200, 600); got != 100 {
		t.Errorf("clampBodyOffset(100, 200, 600) = %d, want 100", got)
	}
	if got := clampBodyOffset(-10, 200, 600); got != 0 {
		t.Errorf("clampBodyOffset(-10, 200, 600) = %d, want 0", got)
	}
	if got := clampBodyOffset(500, 200, 600); got != 400 {
		t.Errorf("clampBodyOffset(500, 200, 600) = %d, want 400 (content-viewport)", got)
	}
	if got := clampBodyOffset(10, 500, 300); got != 0 {
		t.Errorf("clampBodyOffset(10, 500, 300) = %d, want 0 (negative max)", got)
	}
}

func TestWheelScrollsBody(t *testing.T) {
	a, err := newApp()
	if err != nil {
		t.Fatal(err)
	}
	ctx := context.Background()
	if err := a.page.Redraw(ctx); err != nil {
		t.Fatal(err)
	}
	// The compact layout fits the LockView canvas, so the body has no
	// overflow: the wheel must still be consumed (the page never pans)
	// while the offset stays clamped at zero.
	handled, err := a.handleWheel(ctx, 720, 400, 0, -1)
	if err != nil {
		t.Fatal(err)
	}
	if !handled {
		t.Errorf("handleWheel below header dy=-1 = false, want true")
	}
	if a.bodyOffset != 0 {
		t.Errorf("bodyOffset = %d with a fitting body, want 0", a.bodyOffset)
	}
	cur := a.bodyOffset
	handled, err = a.handleWheel(ctx, 720, 10, 0, -1)
	if err != nil {
		t.Fatal(err)
	}
	if handled {
		t.Errorf("handleWheel over header (y=10) = true, want false")
	}
	if a.bodyOffset != cur {
		t.Errorf("bodyOffset = %d after header wheel, want unchanged %d", a.bodyOffset, cur)
	}
}

func TestHeaderStaysStatic(t *testing.T) {
	a, err := newApp()
	if err != nil {
		t.Fatal(err)
	}
	ctx := context.Background()
	if err := a.page.Redraw(ctx); err != nil {
		t.Fatal(err)
	}
	beforeY, beforeH := findInventoryTopbar(t, a)
	for i := 0; i < 20 && a.bodyOffset < 300; i++ {
		if _, err := a.handleWheel(ctx, 720, 400, 0, -1); err != nil {
			t.Fatal(err)
		}
	}
	if err := a.page.Redraw(ctx); err != nil {
		t.Fatal(err)
	}
	if ox, oy := a.page.Scroll().ScrollOffset(); ox != 0 || oy != 0 {
		t.Errorf("ScrollOffset = (%d, %d), want (0, 0): page must never pan", ox, oy)
	}
	afterY, afterH := findInventoryTopbar(t, a)
	if math.Abs(afterY) > 2 {
		t.Errorf("header Y = %v after scroll, want Y≈0 (was %v)", afterY, beforeY)
	}
	if math.Abs(afterY-beforeY) > 1 {
		t.Errorf("header Y moved from %v to %v", beforeY, afterY)
	}
	if math.Abs(afterH-beforeH) > 1 {
		t.Errorf("header H changed from %v to %v", beforeH, afterH)
	}
}

// findInventoryTopbar returns the Y and H of the full-width H≈60 box at Y≈0.
func findInventoryTopbar(t *testing.T, a *app) (y, h float64) {
	t.Helper()
	boxes := a.page.Boxes()
	best := -1
	bestW := 0.0
	for i, b := range boxes {
		if b.H < 55 || b.H > 65 {
			continue
		}
		if math.Abs(b.Y) > 2 {
			continue
		}
		if b.W < 1000 {
			continue
		}
		if b.W > bestW {
			bestW = b.W
			best = i
		}
	}
	if best < 0 {
		t.Fatalf("no full-width H≈60 box at Y≈0 in %d boxes", len(boxes))
	}
	return boxes[best].Y, boxes[best].H
}
