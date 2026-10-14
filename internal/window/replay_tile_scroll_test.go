package window

import (
	"testing"
)

// TestScrollMovesBlitsWithoutRepainting checks a scroll frame repaints no
// tile and copies the whole content at the new offset instead.
func TestScrollMovesBlitsWithoutRepainting(t *testing.T) {
	t.Parallel()

	screen := &dirtyScreen{fakeScreen: &fakeScreen{width: 600, height: 300}}
	screen.display = partialDisplay(600, 300)
	screen.gen = 1
	s := newDirtyShell(screen)
	dst := partialBuf(600, 300)

	s.Draw(dst)

	grid := s.partial.tiles
	if grid == nil {
		t.Fatal("no grid after the first frame")
	}

	s.scrollY = 20
	screen.ok = false
	s.Draw(dst)

	if s.partial.mode != repaintBlit {
		t.Fatalf("scroll mode = %v, want blit", s.partial.mode)
	}

	if !s.partial.whole {
		t.Fatal("scroll frame did not mark the whole content")
	}

	if s.partial.paints != 0 {
		t.Fatalf("scroll repainted %d tiles, want 0", s.partial.paints)
	}

	if s.partial.tiles != grid {
		t.Fatal("scroll rebuilt the grid")
	}

	if full, rects := blitPlan(s.partial.plans, s.partial.chrome); full || len(rects) != 0 {
		t.Fatalf("blit plan = %v (full = %v), want the whole copy", rects, full)
	}
}
