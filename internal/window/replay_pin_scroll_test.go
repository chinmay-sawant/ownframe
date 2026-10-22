package window

import (
	"testing"
)

func TestScrollKeepsPinnedChrome(t *testing.T) {
	t.Parallel()

	screen, s := newPinShell()
	dst := partialBuf(100, 200)

	s.drawReplayPartial(dst)
	grid := s.partial.tiles
	if grid == nil {
		t.Fatal("no grid")
	}

	s.scrollY = 40
	screen.ok = false
	s.drawReplayPartial(dst)

	if s.partial.tiles != grid {
		t.Fatal("scroll rebuilt the grid")
	}

	if s.partial.paints != 0 || !s.partial.whole {
		t.Fatalf("scroll paints = %d whole = %v, want blit", s.partial.paints, s.partial.whole)
	}

	if s.partial.pins == 0 {
		t.Fatal("scroll skipped pinned chrome")
	}

	if s.fullReplay() || !s.hasViewportOps() {
		t.Fatal("scroll with pins must stay on the tile path")
	}
}
