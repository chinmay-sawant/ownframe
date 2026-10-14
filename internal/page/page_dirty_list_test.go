package page

import (
	"image"
	"testing"
)

func TestMarkRectCoalesces(t *testing.T) {
	t.Parallel()

	p := &Page{}
	p.markRect(image.Rect(0, 0, 10, 10))
	p.markRect(image.Rect(14, 0, 24, 10))

	if len(p.dirtyRects) != 1 || p.dirtyRects[0] != image.Rect(0, 0, 24, 10) {
		t.Fatalf("close rects %v", p.dirtyRects)
	}

	p.markRect(image.Rect(80, 80, 90, 90))
	if len(p.dirtyRects) != 2 {
		t.Fatalf("far rect merged: %v", p.dirtyRects)
	}
}

func TestMarkRectCapCollapses(t *testing.T) {
	t.Parallel()

	p := &Page{}
	for i := 0; i <= maxDirtyRects; i++ {
		x := i * 1000
		p.markRect(image.Rect(x, 0, x+10, 10))
	}

	if len(p.dirtyRects) != 1 {
		t.Fatalf("cap left %d rects", len(p.dirtyRects))
	}

	want := image.Rect(0, 0, maxDirtyRects*1000+10, 10)
	if p.dirtyRects[0] != want {
		t.Fatalf("collapse %v, want %v", p.dirtyRects[0], want)
	}
}

func TestTakeDirtyRectsPadClampDrop(t *testing.T) {
	t.Parallel()

	p := &Page{width: 100, height: 100}
	p.markRect(image.Rect(-10, -10, 20, 20))
	p.markRect(image.Rect(200, 200, 210, 210))

	rects, ok := p.TakeDirtyRects()
	if !ok || len(rects) != 1 || rects[0] != image.Rect(0, 0, 24, 24) {
		t.Fatalf("rects %v ok %v", rects, ok)
	}

	if p.dirtyFull || len(p.dirtyRects) != 0 {
		t.Fatalf("take left full=%v rects=%v", p.dirtyFull, p.dirtyRects)
	}
}

func TestTakeDirtyRectsThirdAreaCollapses(t *testing.T) {
	t.Parallel()

	p := &Page{width: 100, height: 100}
	p.markRect(image.Rect(0, 0, 40, 40))
	p.markRect(image.Rect(60, 0, 100, 40))

	rects, ok := p.TakeDirtyRects()
	if !ok || len(rects) != 1 || rects[0] != image.Rect(0, 0, 100, 44) {
		t.Fatalf("rects %v ok %v", rects, ok)
	}
}
