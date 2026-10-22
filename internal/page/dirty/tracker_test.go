package dirty

import (
	"image"
	"testing"
)

// TestTrackerCollapsesPastThird drives the tracker with no page: two
// padded rects covering over a third of the content collapse to one box,
// and the take drains so the next one reports nothing changed.
func TestTrackerCollapsesPastThird(t *testing.T) {
	t.Parallel()

	var d Tracker
	d.MarkRect(image.Rect(0, 0, 40, 40))
	d.MarkRect(image.Rect(60, 0, 100, 40))

	rects, ok := d.Take(image.Rect(0, 0, 100, 100))
	if !ok || len(rects) != 1 {
		t.Fatalf("rects %v ok %v", rects, ok)
	}

	if rects[0] != image.Rect(0, 0, 100, 44) {
		t.Fatalf("collapse %v", rects[0])
	}

	if _, ok := d.Take(image.Rect(0, 0, 100, 100)); ok {
		t.Fatal("take did not drain")
	}
}

// TestTrackerFullReturnsContent checks the whole-frame flag path with
// no page fixture.
func TestTrackerFullReturnsContent(t *testing.T) {
	t.Parallel()

	var d Tracker
	d.MarkRect(image.Rect(0, 0, 10, 10))
	d.MarkFull()

	rects, ok := d.Take(image.Rect(0, 0, 100, 100))
	if !ok || len(rects) != 1 || rects[0] != image.Rect(0, 0, 100, 100) {
		t.Fatalf("rects %v ok %v", rects, ok)
	}
}
