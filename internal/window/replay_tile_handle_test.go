package window

import "testing"

// TestDroppedHandleForcesARebuild pins the retained-content handle contract:
// a caller that drops it (a screenshot, a GPU reset, the gated frame test)
// invalidates the tiles, so the next frame rebuilds them instead of blitting
// pixels nobody can vouch for.
func TestDroppedHandleForcesARebuild(t *testing.T) {
	t.Parallel()

	screen := &dirtyScreen{fakeScreen: &fakeScreen{width: 600, height: 300}}
	screen.display = partialDisplay(600, 300)
	screen.gen = 1
	s := newDirtyShell(screen)
	dst := partialBuf(600, 300)

	s.drawReplayPartial(dst)
	screen.ok = false
	before := s.partial.tiles

	if s.partial.buf == nil {
		t.Fatal("no retained-content handle after a rebuild")
	}

	s.partial.buf = nil
	s.drawReplayPartial(dst)

	if s.partial.tiles == nil || s.partial.tiles == before {
		t.Fatal("dropped handle did not rebuild the grid")
	}

	if s.partial.mode != repaintBuffer {
		t.Fatalf("mode = %v, want a rebuild", s.partial.mode)
	}

	if s.partial.paints != s.partial.tiles.cols*s.partial.tiles.rows {
		t.Fatalf("painted %d tiles, want the whole new grid", s.partial.paints)
	}
}
