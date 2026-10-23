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

func TestTakeAtOneThirdKeepsSeparateRegions(t *testing.T) {
	// The public four-pixel padding produces two 200x150 regions, exactly
	// one third of this 600x300 content. One extra column must collapse them.
	for _, extra := range []int{0, 1} {
		var tracker Tracker
		tracker.MarkRect(image.Rect(4, 4, 196, 146))
		tracker.MarkRect(image.Rect(304, 4, 496+extra, 146))
		rects, ok := tracker.Take(image.Rect(0, 0, 600, 300))
		want := 2
		if extra == 1 {
			want = 1
		}
		if !ok || len(rects) != want {
			t.Fatalf("extra=%d: got %v, want %d regions", extra, rects, want)
		}
		if extra == 0 && (rects[0] != image.Rect(0, 0, 200, 150) || rects[1] != image.Rect(300, 0, 500, 150)) {
			t.Fatalf("padding changed: %v", rects)
		}
	}
}
