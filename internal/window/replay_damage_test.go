package window

import (
	"context"
	"image"
	"testing"
)

// Two separated changes in one frame repaint as two rects, not their union.
func TestTwoRectsRepaintSeparately(t *testing.T) {
	t.Parallel()

	screen := &dirtyScreen{fakeScreen: &fakeScreen{width: 100, height: 100}}
	screen.display = partialDisplay(100, 100)
	screen.gen = 1
	s := newDirtyShell(screen)
	dst := partialBuf(100, 100)

	s.drawReplayPartial(dst)

	first, second := image.Rect(10, 10, 30, 30), image.Rect(60, 60, 80, 80)
	screen.layouts++
	screen.rects, screen.ok = []image.Rectangle{first, second}, true
	screen.gen++
	s.drawReplayPartial(dst)

	if s.partial.mode != repaintRect || len(s.partial.last) != 2 ||
		s.partial.last[0] != first || s.partial.last[1] != second {
		t.Fatalf("mode = %v, rects = %v, want %v plus %v",
			s.partial.mode, s.partial.last, first, second)
	}
}

// A page with only TakeDirty still lands on the per-rect path: the window
// wraps its one rect as a one-element list.
func TestSingleRectTakerStillRepaints(t *testing.T) {
	t.Parallel()

	screen := &oldScreen{fakeScreen: &fakeScreen{width: 100, height: 100}}
	screen.display = partialDisplay(100, 100)
	screen.gen = 1
	s := &shell{
		app: screen, ctx: context.Background(),
		display: screen.display, seq: screen.gen,
		screenW: screen.width, screenH: screen.height,
	}
	dst := partialBuf(100, 100)

	s.drawReplayPartial(dst)

	if s.partial.buf == nil {
		t.Fatal("no buffer for a single-rect page")
	}

	rect := image.Rect(10, 10, 30, 30)
	screen.gen++
	screen.rect, screen.ok = rect, true
	s.drawReplayPartial(dst)

	if screen.takes != 2 {
		t.Fatalf("TakeDirty calls = %d, want 2", screen.takes)
	}

	if s.partial.mode != repaintRect || len(s.partial.last) != 1 ||
		s.partial.last[0] != rect {
		t.Fatalf("mode = %v, rects = %v, want %v", s.partial.mode, s.partial.last, rect)
	}
}
