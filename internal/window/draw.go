package window

import (
	"time"

	"github.com/hajimehoshi/ebiten/v2"
)

func (s *shell) Draw(screen *ebiten.Image) {
	s.devFrameTime()
	defer s.perfFrameSample()
	s.frameBegin()
	defer s.frameCommit()

	if s.fallback {
		defer s.drawBadge(screen)
	}

	start := time.Now()
	s.drawContent(screen)
	s.drawScrollbars(screen)
	s.dev.draw = time.Since(start)
	if s.perf {
		s.dev.stages.draw = s.dev.draw
	}

	tail := time.Now()
	s.drawMenu(screen)

	if s.dev.on {
		s.drawDevTools(screen)
	}
	if s.perf {
		s.dev.stages.present = time.Since(tail)
	}
}

// drawContent paints the page and reports whether the whole screen is
// painted, which every branch does: the tiled replay blits its own tiles, so
// the caller owes no buffer copy. Each full branch lays the page background
// down first, because the screen keeps the previous frame.
func (s *shell) drawContent(screen *ebiten.Image) bool {
	if s.viewLocked() {
		s.partial.live = false
		screen.Fill(s.pageBackground())
		s.drawFitted(screen)

		return true
	}

	if s.bitmapView.img != nil {
		s.partial.live = false
		s.dropDisplayBuffers()
		screen.Fill(s.pageBackground())
		screen.DrawImage(s.bitmapView.img, nil)

		return true
	}

	zoom := s.zoom()

	if s.display != nil {
		if s.stretched() || zoom != 1 {
			s.partial.live = false
			screen.Fill(s.pageBackground())
			s.drawReplayScaled(screen)

			return true
		}

		return s.drawReplayPartial(screen)
	}

	s.partial.live = false
	screen.Fill(s.pageBackground())
	s.drawBitmapPage(screen, zoom)

	return true
}
