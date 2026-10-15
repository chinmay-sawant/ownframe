package page

import (
	"image"
	"testing"
)

// TestDirtyTrackerCollapsesPastThird drives the tracker with no page: two
// padded rects covering over a third of the content collapse to one box,
// and the take drains so the next one reports nothing changed.
func TestDirtyTrackerCollapsesPastThird(t *testing.T) {
	t.Parallel()

	var d dirtyTracker
	d.markRect(image.Rect(0, 0, 40, 40))
	d.markRect(image.Rect(60, 0, 100, 40))

	rects, ok := d.take(image.Rect(0, 0, 100, 100))
	if !ok || len(rects) != 1 {
		t.Fatalf("rects %v ok %v", rects, ok)
	}

	if rects[0] != image.Rect(0, 0, 100, 44) {
		t.Fatalf("collapse %v", rects[0])
	}

	if _, ok := d.take(image.Rect(0, 0, 100, 100)); ok {
		t.Fatal("take did not drain")
	}
}

// TestDirtyTrackerFullReturnsContent checks the whole-frame flag path with
// no page fixture.
func TestDirtyTrackerFullReturnsContent(t *testing.T) {
	t.Parallel()

	var d dirtyTracker
	d.markRect(image.Rect(0, 0, 10, 10))
	d.markFull()

	rects, ok := d.take(image.Rect(0, 0, 100, 100))
	if !ok || len(rects) != 1 || rects[0] != image.Rect(0, 0, 100, 100) {
		t.Fatalf("rects %v ok %v", rects, ok)
	}
}
