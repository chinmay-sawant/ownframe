package window

import (
	"image"
	"testing"
)

func TestClickRepaintsOneRect(t *testing.T) {
	t.Parallel()

	screen := &dirtyScreen{fakeScreen: &fakeScreen{width: 100, height: 100}}
	screen.display = partialDisplay(100, 100)
	screen.gen = 1
	s := newDirtyShell(screen)
	dst := partialBuf(100, 100)

	s.drawReplayPartial(dst)

	screen.click(image.Rect(10, 10, 30, 30))
	s.drawReplayPartial(dst)

	if screen.layouts != 1 {
		t.Fatalf("layouts = %d, want 1", screen.layouts)
	}

	if s.partial.mode != repaintRect || len(s.partial.last) != 1 ||
		s.partial.last[0] != image.Rect(10, 10, 30, 30) {
		t.Fatalf("mode = %v, rects = %v", s.partial.mode, s.partial.last)
	}

	screen.ok = false
	s.drawReplayPartial(dst)

	if s.partial.mode != repaintBlit {
		t.Fatalf("idle mode = %v, want blit", s.partial.mode)
	}

	if screen.takes != 3 {
		t.Fatalf("damage calls = %d, want 3", screen.takes)
	}
}

func TestRemovedOpRepaintsOldBounds(t *testing.T) {
	t.Parallel()

	screen := &dirtyScreen{fakeScreen: &fakeScreen{width: 100, height: 100}}
	screen.display = partialDisplay(100, 100)
	screen.gen = 1
	s := newDirtyShell(screen)
	dst := partialBuf(100, 100)

	s.drawReplayPartial(dst)

	// The box vanishes. The page's diff reports its old bounds, and the
	// window must repaint exactly those so the old pixels are cleared.
	gone := partialDisplay(100, 100)
	gone.Ops = gone.Ops[:1]
	gone.Order = []int{0}
	screen.display, s.display = gone, gone
	screen.click(image.Rect(9, 9, 31, 31))
	s.drawReplayPartial(dst)

	if s.partial.mode != repaintRect || len(s.partial.last) != 1 ||
		s.partial.last[0] != image.Rect(9, 9, 31, 31) {
		t.Fatalf("mode = %v, rects = %v", s.partial.mode, s.partial.last)
	}
}
