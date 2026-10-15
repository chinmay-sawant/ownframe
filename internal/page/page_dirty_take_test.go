package page

import (
	"fmt"
	"image"
	"testing"
)

func TestTakeDirtyIsBoundingBox(t *testing.T) {
	t.Parallel()

	p := &Page{pageGeom: pageGeom{width: 100, height: 100}}
	p.dirty.markRect(image.Rect(0, 0, 10, 10))
	p.dirty.markRect(image.Rect(60, 60, 70, 70))

	rects, ok := p.TakeDirtyRects()
	if !ok || len(rects) != 2 {
		t.Fatalf("rects %v ok %v", rects, ok)
	}

	p.dirty.markRect(image.Rect(0, 0, 10, 10))
	p.dirty.markRect(image.Rect(60, 60, 70, 70))

	r, ok := p.TakeDirty()
	if !ok || r != boundingBox(rects) {
		t.Fatalf("take %v ok %v, want %v", r, ok, boundingBox(rects))
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

	p.dirty.markFull()
	rects, ok := p.TakeDirtyRects()
	if !ok || len(rects) != 1 || rects[0] != image.Rect(0, 0, 100, 100) {
		t.Fatalf("full %v ok %v", rects, ok)
	}

	if p.dirty.full || len(p.dirty.rects) != 0 {
		t.Fatalf("full take left full=%v rects=%v", p.dirty.full, p.dirty.rects)
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

	if p.dirty.pending == nil || len(p.dirty.pending) != 64 {
		t.Fatalf("pending %d ids", len(p.dirty.pending))
	}

	p.TakeDirtyRects()

	if p.dirty.pending == nil {
		t.Fatal("take dropped the pending map")
	}

	if len(p.dirty.pending) != 0 {
		t.Fatalf("pending holds %d ids after take", len(p.dirty.pending))
	}

	p.markPending("id0")
	p.applyPending()

	if p.dirty.pending == nil || len(p.dirty.pending) != 0 {
		t.Fatalf("applyPending left %d ids", len(p.dirty.pending))
	}
}
