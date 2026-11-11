package window

import (
	"github.com/chinmay-sawant/blinkless/layout"
	"github.com/chinmay-sawant/ownframe/internal/host"
	"github.com/hajimehoshi/ebiten/v2"
)

func (s *shell) syncStickyViewport() {
	if !s.viewportSticky {
		return
	}
	observer, ok := s.app.(host.ScrollViewportObserver)
	if !ok {
		return
	}
	width, height := s.viewSize()
	zoom := s.zoom()
	observer.SetScrollViewport(float64(s.scrollX)/zoom, float64(s.scrollY)/zoom, float64(width)/zoom, float64(height)/zoom)
}

// Sticky paint changes with scroll, so its display bypasses cached page tiles.
// Replaying the current viewport clears the previous sticky position as well
// as the new one; ordinary pages retain their existing partial repaint path.
func (s *shell) drawStickyContent(screen *ebiten.Image) {
	s.syncStickyViewport()
	width, height := s.viewSize()
	zoom := s.zoom()
	display := layout.StickyView(s.display, float64(s.scrollX)/zoom, float64(s.scrollY)/zoom, float64(width)/zoom, float64(height)/zoom)
	if display != s.display {
		defer s.renderer.EvictDisplay(display)
	}
	s.partial.live = false
	s.dropDisplayBuffers()
	screen.Fill(s.pageBackground())
	if zoom == 1 && !s.stretched() {
		s.directReplay(screen, display)
		return
	}
	source := s.display
	s.display = display
	s.drawReplayScaled(screen)
	s.display = source
}
