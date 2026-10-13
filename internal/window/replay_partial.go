package window

import (
	"image"

	"github.com/hajimehoshi/ebiten/v2"
)

// drawReplayPartial draws a display-list page through the tile grid when the
// page can report dirty rects, and replays the whole list when it cannot. A
// dirty rect repaints only the tiles it covers and the frame blits only the
// damaged tiles. It reports whether it painted the whole screen, which the
// tiled path does: the retained screen already holds every untouched tile.
func (s *shell) drawReplayPartial(dst *ebiten.Image) bool {
	display := s.display
	if display == nil {
		return true
	}
	s.dropRetainedTiles()

	scroll := image.Pt(s.scrollX, s.scrollY)
	whole := scroll != s.partial.scroll
	s.partial.scroll = scroll
	stale := s.frameStale(dst) || !s.partial.live
	s.partial.live = true
	taker, canTake := s.damageTaker()
	contentW, contentH := s.contentSize()
	if !canTake || s.fullReplay() || !s.bufferOrderSafe() || s.hasViewportOps() {
		s.disposeViewport()
		s.paintFullReplay(dst, display)

		return true
	}

	if oversized(contentW, contentH) {
		// drawViewport drops the tile grid; keep the frame clock.
		focus, size, held := s.partial.focus, s.partial.screen, s.partial.scroll
		s.drawViewport(dst, display, taker)
		s.partial.focus, s.partial.screen, s.partial.scroll = focus, size, held

		return true
	}
	s.disposeViewport()

	rects, ok := taker.TakeDirtyRects()
	plans := planRepaints(contentW, contentH, s.partial, s.app.Generation(), rects, ok)
	s.partial.mode = plans[0].mode
	s.partial.last = s.partial.last[:0]
	s.partial.paints = 0

	for _, plan := range plans {
		s.applyRepaint(display, plan)
		s.partial.last = append(s.partial.last, plan.rect)
	}

	chrome := chromeRects(s.frameRestore(), scroll)
	s.partial.plans, s.partial.chrome, s.partial.whole = plans, chrome, whole
	s.blitFrame(dst, display, plans, chrome, whole, stale)

	return true
}
