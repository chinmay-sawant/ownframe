package main

import (
	"context"
	"sort"
	"testing"

	"github.com/chinmay-sawant/ownframe"
)

// Design canvas for the inventory dashboard. LockView scales this canvas to
// any window, so the layout must fit it with no page scroll.
const (
	inventoryDesignW = 1440.0
	inventoryDesignH = 960.0
	// inventoryFitSlop forgives 2px of border rounding on the canvas edge.
	inventoryFitSlop = 2.0
)

// TestFitsViewport proves the dashboard needs no vertical (or horizontal)
// scroll at the 1440x960 design size: every laid-out box stays inside the
// canvas and the page scroll offset stays at the origin.
func TestFitsViewport(t *testing.T) {
	a, err := newApp()
	if err != nil {
		t.Fatal(err)
	}
	ctx := context.Background()
	if err := a.page.Redraw(ctx); err != nil {
		t.Fatal(err)
	}

	maxX := inventoryDesignW + inventoryFitSlop
	maxY := inventoryDesignH + inventoryFitSlop
	var overflow []string
	for _, b := range a.page.Boxes() {
		if b.W <= 0 {
			continue
		}
		if right := b.X + b.W; right > maxX {
			overflow = append(overflow, "right "+describeBox(b))
		}
		if bottom := b.Y + b.H; bottom > maxY {
			overflow = append(overflow, "bottom "+describeBox(b))
		}
	}
	if len(overflow) > 0 {
		for _, line := range overflow {
			t.Logf("overflow: %s", line)
		}
		logTallestBottoms(t, a)
		t.Errorf("%d box edge(s) exceed the %.0fx%.0f canvas (+%.0fpx tolerance)", len(overflow), inventoryDesignW, inventoryDesignH, inventoryFitSlop)
	}

	if ox, oy := a.page.Scroll().ScrollOffset(); ox != 0 || oy != 0 {
		t.Errorf("ScrollOffset = (%d, %d), want (0, 0): design canvas must not scroll", ox, oy)
	}
}

// TestLockViewNoZoom is the headless proxy for the LockView no-zoom policy:
// driving the wheel hard over the content must stay clamped inside the inner
// body viewport and must never pan the page itself. True zoom-chord testing
// (ctrl+wheel / pinch in a live window) needs a window host, so config plus
// this clamp is the headless stand-in.
func TestLockViewNoZoom(t *testing.T) {
	a, err := newApp()
	if err != nil {
		t.Fatal(err)
	}
	ctx := context.Background()
	if err := a.page.Redraw(ctx); err != nil {
		t.Fatal(err)
	}

	for i := 0; i < 20; i++ {
		if _, err := a.handleWheel(ctx, 720, 400, 0, -1); err != nil {
			t.Fatalf("handleWheel #%d: %v", i, err)
		}
	}
	if a.bodyOffset < 0 {
		t.Errorf("bodyOffset = %d after 20x wheel down, want >= 0", a.bodyOffset)
	}
	viewportH, contentH := a.bodyMetrics()
	if want := clampBodyOffset(a.bodyOffset, viewportH, contentH); a.bodyOffset != want {
		t.Errorf("bodyOffset = %d, want clamped %d (viewport %d, content %d)", a.bodyOffset, want, viewportH, contentH)
	}
	if ox, oy := a.page.Scroll().ScrollOffset(); ox != 0 || oy != 0 {
		t.Errorf("ScrollOffset = (%d, %d) after wheel storm, want (0, 0): page must never pan", ox, oy)
	}
}

// describeBox renders one overflow line for failure output.
func describeBox(b ownframe.Box) string {
	return "tag=" + b.Tag + " id=" + b.ID + " action=" + b.Action
}

// logTallestBottoms logs the five boxes with the largest Y+H so an overflow
// failure points at the tallest offenders first.
func logTallestBottoms(t *testing.T, a *app) {
	t.Helper()
	boxes := a.page.Boxes()
	sort.Slice(boxes, func(i, j int) bool {
		return boxes[i].Y+boxes[i].H > boxes[j].Y+boxes[j].H
	})
	for i := 0; i < 5 && i < len(boxes); i++ {
		b := boxes[i]
		t.Logf("tallest[%d]: tag=%s id=%q x=%.1f y=%.1f w=%.1f h=%.1f bottom=%.1f", i, b.Tag, b.ID, b.X, b.Y, b.W, b.H, b.Y+b.H)
	}
}
