package window

import (
	"time"

	"github.com/hajimehoshi/ebiten/v2"
)

// drawDevTools paints the overlay over the frame: the operation outlines,
// the chosen operation, the hovered and pinned boxes, and the right dock.
// It only reads shell state, so the page picture never changes.
func (s *shell) drawDevTools(screen *ebiten.Image) {
	refresh := s.refreshDevPanel(time.Now())

	if s.dev.ops {
		s.drawDevOps(screen)
	}

	if s.dev.haveOp {
		s.drawDevOpPick(screen)
	}

	s.drawDevPick(screen)
	s.drawCachedDevPanel(screen, refresh)
}
