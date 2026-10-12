package window

import (
	"image"

	"github.com/hajimehoshi/ebiten/v2"
)

// drawReplayPartial draws a display-list page through the persistent buffer
// when the page can report dirty rects, and replays the whole list when it
// cannot. It reports whether it painted the whole screen; otherwise the
// frame's screen blits wait in partialState for the caller.
func (s *shell) drawReplayPartial(dst *ebiten.Image) bool {
	display := s.display
	if display == nil {
		return true
	}

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
		// drawViewport drops the partial buffer; keep the frame clock.
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

	for _, plan := range plans {
		s.applyRepaint(display, plan)
		s.partial.last = append(s.partial.last, plan.rect)
	}

	chrome := chromeRects(s.frameRestore(), scroll)
	if stale || s.partial.buf == nil || !withinBuffer(s.partial.buf, chrome) {
		// applyRepaint above brought the buffer up to date for this frame, so
		// show it whole instead of replaying the list a second time. A nil
		// buffer still needs the direct replay.
		if s.partial.buf != nil {
			s.blitWhole(dst)
		} else {
			s.paintFullReplay(dst, display)
		}

		return true
	}

	if full, _ := blitPlan(plans, chrome); full {
		if s.partial.buf != nil {
			s.blitWhole(dst)
		} else {
			s.paintFullReplay(dst, display)
		}

		return true
	}

	s.partial.plans, s.partial.chrome, s.partial.whole = plans, chrome, whole

	return false
}
