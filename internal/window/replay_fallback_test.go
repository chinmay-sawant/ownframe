package window

import "testing"

func TestPartialPathNeedsATaker(t *testing.T) {
	t.Parallel()

	plain := &fakeScreen{width: 100, height: 100}
	s := &shell{
		app: plain, display: partialDisplay(100, 100), seq: 1,
		screenW: 100, screenH: 100,
	}

	s.drawReplayPartial(partialBuf(100, 100))

	if s.partial.tiles != nil {
		t.Fatal("grid built without a dirty taker")
	}
}

func TestTickingPageKeepsFullReplay(t *testing.T) {
	t.Parallel()

	screen := &dirtyScreen{fakeScreen: &fakeScreen{width: 100, height: 100}}
	screen.display = partialDisplay(100, 100)
	screen.ticking = true
	s := newDirtyShell(screen)

	s.drawReplayPartial(partialBuf(100, 100))

	if screen.takes != 0 {
		t.Fatalf("damage calls = %d, want 0", screen.takes)
	}

	if s.partial.tiles != nil {
		t.Fatal("grid built for a ticking page")
	}
}

// TestScrollOnlyFrameBlitsWholeBuffer checks that a committed frame leaves a
// scroll-only frame one whole-buffer copy at the new offset: the repaintBlit
// plan contributes no rects, so the whole flag supplies the copy.
func TestScrollOnlyFrameBlitsWholeBuffer(t *testing.T) {
	t.Parallel()

	screen := &dirtyScreen{fakeScreen: &fakeScreen{width: 100, height: 100}}
	screen.display = partialDisplay(100, 100)
	screen.gen = 1
	s := newDirtyShell(screen)
	dst := partialBuf(100, 100)

	s.Draw(dst)
	s.scrollY = 20
	screen.ok = false
	s.Draw(dst)

	if s.partial.mode != repaintBlit {
		t.Fatalf("scroll mode = %v, want blit", s.partial.mode)
	}

	if !s.partial.whole {
		t.Fatal("scroll frame did not set whole")
	}

	if full, rects := blitPlan(s.partial.plans, s.partial.chrome); full || len(rects) != 0 {
		t.Fatalf("blit plan = %v (full = %v), want no rects; whole supplies the copy", rects, full)
	}
}
