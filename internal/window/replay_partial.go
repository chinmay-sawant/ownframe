package window

import (
	"image"

	"github.com/chinmay-sawant/blinkless/layout"
	"github.com/chinmay-sawant/ownframe/internal/replay"
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
		// The viewport path drops the tile grid. A later stable blit may
		// idle. A scroll, a generation change, a dirty region, a focus or
		// size stale frame, or a ticker keeps the frame clock.
		focus, size, held := s.partial.focus, s.partial.screen, s.partial.scroll
		s.drawViewport(dst, display, taker)
		s.partial.focus, s.partial.screen, s.partial.scroll = focus, size, held
		if stale {
			s.viewport.blitOnly = false
		}

		return true
	}
	s.disposeViewport()
	// Prepare this frame's retained layers before planning, so their old ∪
	// new damage joins the page's dirty rects and the tile pass can skip their
	// owned operations.
	layerDamage := s.renderer.BeginLayers(display, s.app.Generation(), s.layerSpecs(display))

	clipX := s.dockClipX(contentW)
	rects, ok := taker.TakeDirtyRects()
	if ok && len(layerDamage) > 0 {
		content := image.Rect(0, 0, contentW, contentH)
		combined := make([]image.Rectangle, 0, len(rects)+len(layerDamage))
		combined = append(combined, rects...)
		for _, rect := range layerDamage {
			if inside := rect.Intersect(content); !inside.Empty() {
				combined = append(combined, inside)
			}
		}
		rects = combined
	}
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

// layerSpecs asks the screen for this frame's promoted elements. A screen
// without the role keeps the direct replay path.
func (s *shell) layerSpecs(d *layout.Display) []replay.LayerSpec {
	source, ok := s.app.(replay.LayerSource)
	if !ok {
		return nil
	}

	return source.LayerSpecs(d)
}
