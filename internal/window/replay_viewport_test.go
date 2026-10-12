package window

import (
	"image"
	"testing"
)

// viewportTaker reports scripted dirty rects for the oversized path.
type viewportTaker struct {
	rects []image.Rectangle
	ok    bool
	takes int
}

func (v *viewportTaker) TakeDirtyRects() ([]image.Rectangle, bool) {
	v.takes++

	return v.rects, v.ok
}

func TestViewportBufferStaysWindowSized(t *testing.T) {
	app := &dirtyScreen{fakeScreen: &fakeScreen{width: 80, height: 60}, gen: 1}
	app.display = partialDisplay(80, 37000)
	s := newDirtyShell(app)
	taker := &viewportTaker{}
	dst := partialBuf(80, 60)
	s.partial.buf = partialBuf(80, 100)
	s.drawViewport(dst, app.display, taker)
	if got := s.viewport.buf.Bounds(); got != image.Rect(0, 0, 80, 180) {
		t.Fatalf("buffer = %v", got)
	}
	if s.partial.buf != nil {
		t.Fatal("content buffer retained")
	}
	saved := s.viewport.buf
	s.drawViewport(dst, app.display, taker)
	if s.viewport.buf != saved {
		t.Fatal("idle frame reallocated")
	}
	s.scrollY = 2000
	s.drawViewport(dst, app.display, taker)
	if s.viewport.area.Min.Y > 2000 || s.viewport.area.Max.Y < 2060 || s.viewport.buf != saved {
		t.Fatal("scroll lost or replaced the buffer")
	}
	app.gen++
	s.seq = app.gen
	s.drawViewport(dst, app.display, taker)
	if s.viewport.key.generation != app.gen {
		t.Fatal("generation did not invalidate")
	}
	if taker.takes != 4 {
		t.Fatalf("takes = %d, want 4", taker.takes)
	}
	s.disposeViewport()
	if s.viewport.buf != nil || s.viewport.valid {
		t.Fatal("viewport not released")
	}
}

func TestViewportDirtyBounds(t *testing.T) {
	one := image.Rect(10, 20, 30, 40)
	if got := viewportDirtyBounds([]image.Rectangle{one}); got != one {
		t.Fatalf("one rect = %v, want %v", got, one)
	}
	two := []image.Rectangle{one, image.Rect(200, 300, 300, 400)}
	if got := viewportDirtyBounds(two); got != image.Rect(10, 20, 300, 400) {
		t.Fatalf("two rects = %v", got)
	}
	if got := viewportDirtyBounds(nil); !got.Empty() {
		t.Fatalf("empty list = %v, want empty", got)
	}
}
