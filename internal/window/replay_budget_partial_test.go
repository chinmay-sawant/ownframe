package window

import (
	"image"
	"testing"

	"github.com/hajimehoshi/ebiten/v2"
)

// TestPartialBudgetKeepsTiledPath checks that an ordinary large page keeps
// the tile grid and reports an idle frame. The page is over the old four
// megabyte budget, so an absolute count would have moved it to the viewport
// path.
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

// TestPartialBudgetViewportIdles checks that a stationary viewport blit can
// park. A scroll inside the overscan, a generation change, a ticker, or a
// dirty rect keeps the frame clock.
func TestPartialBudgetViewportIdles(t *testing.T) {
	t.Parallel()

	// Desktop Run turns clearing off. A host that clears every frame marks
	// the blit stale, so this test matches that desktop setting.
	ebiten.SetScreenClearedEveryFrame(false)

	screen := &dirtyScreen{fakeScreen: &fakeScreen{width: 1920, height: 1080}}
	screen.display = partialDisplay(1920, 22000)
	screen.gen = 1
	s := newDirtyShell(screen)
	dst := partialBuf(1920, 1080)

	s.drawReplayPartial(dst)
	s.frameCommit()
	if s.viewport.buf == nil {
		t.Fatal("a huge page kept the content grid")
	}
	if s.idleCleanFrame() {
		t.Fatal("the first viewport frame must read dirty")
	}

	screen.ok = false
	settleViewportWarmup(s, dst)
	if !s.idleCleanFrame() {
		t.Fatal("a stable viewport blit must read clean")
	}

	s.scrollY = 8
	s.drawReplayPartial(dst)
	s.frameCommit()
	if s.idleCleanFrame() {
		t.Fatal("a scroll inside the overscan must read dirty")
	}

	s.drawReplayPartial(dst)
	s.frameCommit()
	if !s.idleCleanFrame() {
		t.Fatal("the same scroll must read clean")
	}

	screen.gen++
	s.seq = screen.gen
	s.drawReplayPartial(dst)
	s.frameCommit()
	if s.idleCleanFrame() {
		t.Fatal("a generation change must read dirty")
	}

	settleViewportWarmup(s, dst)
	if !s.idleCleanFrame() {
		t.Fatal("the new generation with no dirty must read clean")
	}

	screen.ticking = true
	s.drawReplayPartial(dst)
	s.frameCommit()
	if s.idleCleanFrame() {
		t.Fatal("a ticking viewport must read dirty")
	}

	screen.ticking = false
	s.drawReplayPartial(dst)
	s.frameCommit()
	if !s.idleCleanFrame() {
		t.Fatal("a quiet viewport must read clean after ticking stops")
	}

	screen.ok = true
	screen.rects = []image.Rectangle{image.Rect(10, 10, 30, 30)}
	s.drawReplayPartial(dst)
	s.frameCommit()
	if s.idleCleanFrame() {
		t.Fatal("a dirty rect must read dirty")
	}
}

// Warmup progresses over frames before a static page can park.
func settleViewportWarmup(s *shell, dst *ebiten.Image) {
	for range 100 {
		s.drawReplayPartial(dst)
		s.frameCommit()
		if s.idleCleanFrame() {
			return
		}
	}
}
