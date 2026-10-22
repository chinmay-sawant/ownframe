package window

import (
	"testing"
)

// TestPartialBudgetKeepsTiledPath checks that an ordinary large page keeps
// the tile grid and reports an idle frame. The page is over the old four
// megabyte budget, so an absolute count would have moved it to the viewport
// path, which never reports an idle frame.
func TestPartialBudgetKeepsTiledPath(t *testing.T) {
	t.Parallel()

	screen := &dirtyScreen{fakeScreen: &fakeScreen{width: 1920, height: 1080}}
	screen.display = partialDisplay(1920, 2200)
	screen.gen = 1
	s := newDirtyShell(screen)
	dst := partialBuf(1920, 1080)

	s.drawReplayPartial(dst)
	s.drawReplayPartial(dst)

	if s.viewport.buf != nil {
		t.Fatal("a large page took the viewport path")
	}
	if s.partial.tiles == nil {
		t.Fatal("no content grid")
	}
	if !s.idleCleanFrame() {
		t.Fatal("a large page never reports an idle frame")
	}
}

// TestPartialBudgetTakesViewportPath checks that a page far bigger than the
// window takes the viewport path, which keeps its memory bounded.
func TestPartialBudgetTakesViewportPath(t *testing.T) {
	t.Parallel()

	screen := &dirtyScreen{fakeScreen: &fakeScreen{width: 1920, height: 1080}}
	screen.display = partialDisplay(1920, 22000)
	screen.gen = 1
	s := newDirtyShell(screen)
	dst := partialBuf(1920, 1080)

	s.drawReplayPartial(dst)

	if s.viewport.buf == nil {
		t.Fatal("a huge page kept the content grid")
	}
	if s.partial.tiles != nil {
		t.Fatal("the content grid was retained")
	}
}
