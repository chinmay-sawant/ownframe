package window

import (
	"github.com/chinmay-sawant/blinkless/layout"
	"github.com/hajimehoshi/ebiten/v2"
)

// applyRepaint brings the tile grid up to date for this frame. A rebuild
// ensures the grid matches the content and repaints all of it; a rect plan
// repaints only the tiles the rect covers. A blit plan changes no pixels.
func (s *shell) applyRepaint(display *layout.Display, plan repaintPlan) {
	switch plan.mode {
	case repaintBuffer:
		s.ensureTiles(plan.rect.Dx(), plan.rect.Dy())
		s.paintTiles(display, plan.rect)
	case repaintRect:
		s.paintTiles(display, plan.rect)
	}

	s.partial.gen = s.app.Generation()
}

// paintFullReplay paints the whole screen: the page background, then the
// display list. It is the full branch of the replay path, where the tiles
// alone cannot supply the frame.
func (s *shell) paintFullReplay(dst *ebiten.Image, display *layout.Display) {
	dst.Fill(s.pageBackground())
	s.directReplay(dst, display)
}
