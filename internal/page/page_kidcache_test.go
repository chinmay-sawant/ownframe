package page

import (
	"image"
	"testing"
)

// kidCacheFlags snapshots the lazy kid index state. The box cache is shared by
// every page in the process, so the read takes its mutex even in a test.
func kidCacheFlags(t *testing.T, p *Page) (ready, staleRects bool) {
	t.Helper()
	boxCaches.Lock()
	defer boxCaches.Unlock()

	e, ok := boxCaches.byPage[p]
	if !ok {
		t.Fatal("page has no box cache")
	}

	return e.kid.ready, e.kid.staleRects
}

// TestKidIndexBuildsOnFirstUse pins the lazy build: a commit only records the
// box list, and the first subtree query builds the parent-to-children index.
func TestKidIndexBuildsOnFirstUse(t *testing.T) {
	p := &Page{}
	p.setBoxes([]Box{{ID: "root", ParentIndex: -1, W: 10, H: 10}})

	if ready, _ := kidCacheFlags(t, p); ready {
		t.Fatal("setBoxes must not build the kid index")
	}

	r, ok := kidRect(p, "root")
	if !ok {
		t.Fatal("kidRect missed the committed box")
	}
	if want := image.Rect(0, 0, 10, 10); r != want {
		t.Fatalf("kidRect = %v, want %v", r, want)
	}

	if ready, _ := kidCacheFlags(t, p); !ready {
		t.Fatal("the first kidRect must build the index")
	}
}

// TestKidIndexUpdatesRectsInPlace keeps a moved subtree on the built index:
// the id sequence and parent links are unchanged, so only the rects snapshot
// is rewritten and a second query answers with the new geometry.
func TestKidIndexUpdatesRectsInPlace(t *testing.T) {
	p := &Page{}
	p.setBoxes([]Box{
		{ID: "card", ParentIndex: -1, X: 0, Y: 0, W: 100, H: 100},
		{ID: "badge", ParentIndex: 0, X: 90, Y: 90, W: 40, H: 40},
	})

	if _, ok := kidRect(p, "card"); !ok {
		t.Fatal("kidRect missed the card")
	}

	p.setBoxes([]Box{
		{ID: "card", ParentIndex: -1, X: 0, Y: 0, W: 100, H: 100},
		{ID: "badge", ParentIndex: 0, X: 200, Y: 200, W: 40, H: 40},
	})

	ready, stale := kidCacheFlags(t, p)
	if !ready || !stale {
		t.Fatalf("same-shape commit kept ready=%v stale=%v, want the index kept and the rects stale", ready, stale)
	}

	r, ok := kidRect(p, "card")
	if !ok {
		t.Fatal("kidRect missed the moved card")
	}
	if want := image.Rect(0, 0, 240, 240); r != want {
		t.Fatalf("kidRect = %v, want the moved subtree %v", r, want)
	}

	if ready, stale := kidCacheFlags(t, p); !ready || stale {
		t.Fatalf("after the rect rewrite ready=%v stale=%v, want a current index", ready, stale)
	}
}

// TestKidIndexRebuildsOnStructuralChange drops the index when a parent link or
// the id sequence changes, so the subtree query follows the new tree.
func TestKidIndexRebuildsOnStructuralChange(t *testing.T) {
	p := &Page{}
	p.setBoxes([]Box{
		{ID: "card", ParentIndex: -1, X: 0, Y: 0, W: 100, H: 100},
		{ID: "badge", ParentIndex: 0, X: 200, Y: 200, W: 40, H: 40},
	})

	if r, ok := kidRect(p, "card"); !ok || r != image.Rect(0, 0, 240, 240) {
		t.Fatalf("built card rect = %v, %v, want the badge inside", r, ok)
	}

	p.setBoxes([]Box{
		{ID: "card", ParentIndex: -1, X: 0, Y: 0, W: 100, H: 100},
		{ID: "badge", ParentIndex: -1, X: 200, Y: 200, W: 40, H: 40},
	})

	if ready, _ := kidCacheFlags(t, p); ready {
		t.Fatal("a parent change must drop the kid index")
	}

	r, ok := kidRect(p, "card")
	if !ok {
		t.Fatal("kidRect missed the recut card")
	}
	if want := image.Rect(0, 0, 100, 100); r != want {
		t.Fatalf("rebuilt card rect = %v, want the badge detached %v", r, want)
	}

	p.setBoxes([]Box{{ID: "card", ParentIndex: -1, X: 0, Y: 0, W: 100, H: 100}})

	if ready, _ := kidCacheFlags(t, p); ready {
		t.Fatal("a removed box must drop the kid index")
	}
}
