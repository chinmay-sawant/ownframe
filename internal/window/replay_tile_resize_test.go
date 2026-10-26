package window

import (
	"image"
	"testing"

	"github.com/chinmay-sawant/blinkless/layout"
)

// TestContentChangeRebuildsTheGrid checks a content size change disposes the
// old grid and allocates one for the new content. Only the visible tiles
// paint; the row below the fold stays dirty until a scroll brings it on
// screen, when the blit repaints it on demand.
func TestContentChangeRebuildsTheGrid(t *testing.T) {
	t.Parallel()

	screen := &dirtyScreen{fakeScreen: &fakeScreen{width: 600, height: 300}}
	screen.display = partialDisplay(600, 300)
	screen.gen = 1
	s := newDirtyShell(screen)

	s.drawReplayPartial(partialBuf(600, 300))
	first := s.partial.tiles

	screen.boxes = []layout.Box{{X: 0, Y: 0, W: 600, H: 700}}
	screen.click(image.Rect(10, 10, 20, 20))
	s.drawReplayPartial(partialBuf(600, 300))

	if s.partial.tiles == nil || s.partial.tiles == first {
		t.Fatal("size change kept the grid")
	}

	if got := s.partial.tiles.content(); got != image.Rect(0, 0, 600, 700) {
		t.Fatalf("new grid covers %v, want the 600x700 content", got)
	}

	if s.partial.paints != 6 {
		t.Fatalf("painted %d tiles, want the 6 visible ones", s.partial.paints)
	}

	s.scrollY = 400
	screen.ok = false
	s.drawReplayPartial(partialBuf(600, 300))

	if s.partial.paints != 3 {
		t.Fatalf("scroll painted %d tiles, want the 3 newly visible ones", s.partial.paints)
	}
}
