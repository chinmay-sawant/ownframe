package page

import (
	"fmt"
	"image"
	"testing"
)

func TestTakeDirtyIsBoundingBox(t *testing.T) {
	t.Parallel()

	p := &Page{pageGeom: pageGeom{width: 100, height: 100}}
	p.dirty.MarkRect(image.Rect(0, 0, 10, 10))
	p.dirty.MarkRect(image.Rect(60, 60, 70, 70))

	rects, ok := p.TakeDirtyRects()
	if !ok || len(rects) != 2 {
		t.Fatalf("rects %v ok %v", rects, ok)
	}

	p.dirty.MarkRect(image.Rect(0, 0, 10, 10))
	p.dirty.MarkRect(image.Rect(60, 60, 70, 70))

	r, ok := p.TakeDirty()
	want := rects[0].Union(rects[1])
	if !ok || r != want {
		t.Fatalf("take %v ok %v, want %v", r, ok, want)
	}
}

func TestTakeDirtyRectsEmptyAndFull(t *testing.T) {
	t.Parallel()

	p := &Page{pageGeom: pageGeom{width: 100, height: 100}}
	if rects, ok := p.TakeDirtyRects(); ok || len(rects) != 0 {
		t.Fatalf("empty %v ok %v", rects, ok)
	}

	if r, ok := p.TakeDirty(); ok || !r.Empty() {
		t.Fatalf("empty take %v ok %v", r, ok)
	}

	p.dirty.MarkFull()
	rects, ok := p.TakeDirtyRects()
	if !ok || len(rects) != 1 || rects[0] != image.Rect(0, 0, 100, 100) {
		t.Fatalf("full %v ok %v", rects, ok)
	}

	if p.dirty.Full || len(p.dirty.Rects) != 0 {
		t.Fatalf("full take left full=%v rects=%v", p.dirty.Full, p.dirty.Rects)
	}

	if _, ok := p.TakeDirtyRects(); ok {
		t.Fatal("full take did not consume")
	}
}

func TestPendingMapReused(t *testing.T) {
	t.Parallel()

	p := &Page{pageGeom: pageGeom{width: 100, height: 100}}
	for i := 0; i < 64; i++ {
		p.markPending(fmt.Sprintf("id%d", i))
	}

	if p.dirty.Pending == nil || len(p.dirty.Pending) != 64 {
		t.Fatalf("pending %d ids", len(p.dirty.Pending))
	}

	p.TakeDirtyRects()

	if p.dirty.Pending == nil {
		t.Fatal("take dropped the pending map")
	}

	if len(p.dirty.Pending) != 0 {
		t.Fatalf("pending holds %d ids after take", len(p.dirty.Pending))
	}

	p.markPending("id0")
	p.applyPending()

	if p.dirty.Pending == nil || len(p.dirty.Pending) != 0 {
		t.Fatalf("applyPending left %d ids", len(p.dirty.Pending))
	}
}
