package window

import (
	"image"

	"github.com/hajimehoshi/ebiten/v2"
)

// drawReplayPartial draws a display-list page through the tile grid when the
// page can report dirty rects, and replays the whole list when it cannot. A
// dirty rect repaints only its tiles and the frame blits only the damaged
// ones, trimmed to the page the devtools dock leaves visible.
func (s *shell) drawReplayPartial(dst *ebiten.Image) bool {
	display := s.display
	if display == nil {
		return true
	}
	s.dropRetainedTiles()
	s.partial.paints, s.partial.paintPx, s.partial.pins = 0, 0, 0
	s.partial.blits, s.partial.blitPx = 0, 0

	scroll := image.Pt(s.scrollX, s.scrollY)
	whole := scroll != s.partial.scroll
	s.partial.scroll = scroll
	stale := s.frameStale(dst) || !s.partial.live
	s.partial.live = true
	taker, canTake := s.damageTaker()
	contentW, contentH := s.contentSize()
	if !canTake || s.fullReplay() || !s.bufferOrderSafe() {
		s.disposeViewport()
		s.paintFullReplay(dst, display)

		return true
	}

	if oversized(contentW, contentH) || s.gridTooBig(contentW, contentH) {
		// A grid that dwarfs the window takes the viewport path, which
		// drops the tile grid; keep the frame clock.
		focus, size, held := s.partial.focus, s.partial.screen, s.partial.scroll
		s.drawViewport(dst, display, taker)
		s.partial.focus, s.partial.screen, s.partial.scroll = focus, size, held

		return true
	}
	s.disposeViewport()

	clipX := s.dockClipX(contentW)
	rects, ok := taker.TakeDirtyRects()
	plans := planRepaints(contentW, contentH, s.partial, s.app.Generation(), rects, ok, clipX)
	s.partial.mode = plans[0].mode
	s.partial.clipX = clipX
	s.partial.last = s.partial.last[:0]

	for _, plan := range plans {
		s.applyRepaint(display, plan)
		s.partial.last = append(s.partial.last, plan.rect)
	}

	chrome := chromeRects(s.frameRestore(), scroll)
	s.partial.plans, s.partial.chrome, s.partial.whole = plans, chrome, whole
	s.blitFrame(dst, display, plans, chrome, whole, stale)

	return true
}
