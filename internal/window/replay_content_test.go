package window

import (
	"testing"

	"github.com/chinmay-sawant/blinkless/layout"
)

// TestBufferCoversContentPastTheCanvas pins the scrolling example: the
// laid-out canvas is the window size, but the page's boxes reach past it.
// The replay buffer and the full plan must cover the content, or scrolling
// paints blank.
func TestBufferCoversContentPastTheCanvas(t *testing.T) {
	t.Parallel()

	screen := &dirtyScreen{fakeScreen: &fakeScreen{width: 100, height: 100}}
	screen.display = partialDisplay(100, 100)
	screen.boxes = []layout.Box{{X: 0, Y: 0, W: 100, H: 400}}
	screen.gen = 1
	s := newDirtyShell(screen)

	s.drawReplayPartial(partialBuf(100, 100))

	if s.partial.buf == nil {
		t.Fatal("no buffer for the taller content")
	}

	if got := s.partial.buf.Bounds().Dy(); got != 400 {
		t.Fatalf("buffer height = %d, want the 400 px content", got)
	}

	if len(s.partial.last) != 1 || s.partial.last[0].Dy() != 400 {
		t.Fatalf("full plan rects = %v, want one 400 px rect", s.partial.last)
	}
}
