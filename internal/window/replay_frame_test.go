package window

import "testing"

// TestFrameDirtyKeepsTheBuffer checks a ticking page that opted into dirty
// rects builds the buffer once and then leaves the retained screen alone: an
// idle frame plans a blit that copies nothing.
func TestFrameDirtyKeepsTheBuffer(t *testing.T) {
	t.Parallel()

	screen := &dirtyScreen{fakeScreen: &fakeScreen{width: 100, height: 100}}
	screen.display = partialDisplay(100, 100)
	screen.gen = 1
	screen.ticking = true
	screen.frame = true
	s := newDirtyShell(screen)
	dst := partialBuf(100, 100)

	// The Update captures the opt-in. The draws below model Ebiten issuing a
	// second Draw without another Update: the captured value must stand.
	s.captureFrameOptIn()

	s.Draw(dst)

	if s.partial.tiles == nil {
		t.Fatal("opted-in tick did not build the tile grid")
	}

	s.Draw(dst)

	if s.partial.mode != repaintBlit {
		t.Fatalf("second frame mode = %v, want blit", s.partial.mode)
	}

	if s.partial.whole {
		t.Fatal("idle frame set whole without a scroll")
	}

	full, rects := blitPlan(s.partial.plans, s.partial.chrome)
	if full || len(rects) != 0 {
		t.Fatalf("idle frame blits = %v (full = %v), want none", rects, full)
	}
}
