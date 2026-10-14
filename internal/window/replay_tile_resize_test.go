package window

import (
	"image"
	"testing"

	"github.com/chinmay-sawant/blinkless/layout"
)

// TestContentChangeRebuildsTheGrid checks a content size change disposes the
// old grid and allocates one for the new content.
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

	if s.partial.paints != s.partial.tiles.cols*s.partial.tiles.rows {
		t.Fatalf("painted %d tiles, want the whole new grid", s.partial.paints)
	}
}
