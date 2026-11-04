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
	s.partial.paints, s.partial.paintPx, s.partial.pins = 0, 0, 0
	s.partial.blits, s.partial.blitPx = 0, 0
	if s.renderer.HasBackdropFilter(display) {
		generation := s.app.Generation()
		scroll := image.Pt(s.scrollX, s.scrollY)
		backdropClean := s.partial.live && s.partial.mode == repaintBlit &&
			!s.partial.whole && s.partial.gen == generation && s.partial.scroll == scroll && !s.frameOptIn
		// Filtered output depends on pixels outside the element's own dirty
		// rect. Replay the whole display until the damage path tracks those
		// backdrop dependencies.
		s.dropTiles()
		s.disposeViewport()
		// Backdrop shaders read the scene painted so far. Build it in a
		// texture, then present once, rather than sampling the window buffer.
		bounds := dst.Bounds()
		frame := s.ensureScaleBuf(bounds.Dx(), bounds.Dy())
		s.paintFullReplay(frame, display)
		op := &ebiten.DrawImageOptions{Blend: ebiten.BlendCopy}
		op.GeoM.Translate(float64(bounds.Min.X), float64(bounds.Min.Y))
		dst.DrawImage(frame, op)
		if taker, ok := s.damageTaker(); ok {
			rects, available := taker.TakeDirtyRects()
			if available && len(rects) > 0 {
				backdropClean = false
			}
		}
		s.partial.gen = generation
		s.partial.scroll = scroll
		s.partial.mode = repaintBlit
		s.partial.whole = false
		s.partial.live = true
		s.partial.backdropClean = backdropClean

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
	if !canTake || s.fullReplay() || !s.bufferOrderSafe() {
		s.disposeViewport()
		s.dropScaleBuf()
		s.paintFullReplay(dst, display)

		return true
	}

	if oversized(contentW, contentH) || s.gridTooBig(contentW, contentH) || s.windowing() {
		// The viewport path drops the tile grid. A later stable blit may
		// idle, including with a quiet tick. A scroll, a generation change,
		// a dirty region, a focus or size stale frame, or a tick that
		// reported damage keeps the frame clock. A windowing page always
		// takes this path, so its working set stays viewport-sized no
		// matter how tall its spacers make the content.
		focus, size, held := s.partial.focus, s.partial.screen, s.partial.scroll
		s.drawViewport(dst, display, taker)
		s.partial.focus, s.partial.screen, s.partial.scroll = focus, size, held
		if stale {
			s.viewport.blitOnly = false
		}

		return true
	}
	s.disposeViewport()
	s.dropScaleBuf()
	// Prepare this frame's retained layers before planning, so their old ∪
	// new damage joins the page's dirty rects and the tile pass can skip their
	// owned operations.
	layerDamage := s.renderer.BeginLayers(display, s.layerGeneration(), s.layerSpecs(display))

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

	chrome := chromeRectsBuf(s.partial.chrome[:0], s.frameRestore(), scroll)
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

// layerGeneration keys the retained layer cache. It is the page's content
// generation, so a re-projection pass that rewrites operation matrices in
// place reuses the layer textures instead of re-baking them every frame. A
// screen without the role falls back to its display generation.
func (s *shell) layerGeneration() uint64 {
	if source, ok := s.app.(layerGenerator); ok {
		return source.LayerGeneration()
	}

	return s.app.Generation()
}

// layerGenerator is the optional screen role that separates display content
// from a re-projection pass.
type layerGenerator interface {
	LayerGeneration() uint64
}

// layerPromoter is the optional screen role that learns which retained layers
// the compositor kept this frame. A promoted element's paint changes land in
// its layer texture, so the page stops damaging the background for it.
type layerPromoter interface {
	SetPromotedLayers(ids []string)
}

// syncPromotedLayers hands the active layer ids back to the page. A demoted
// layer leaves the set, so its next sample damages the background again.
func (s *shell) syncPromotedLayers() {
	promoter, ok := s.app.(layerPromoter)
	if !ok {
		return
	}

	promoter.SetPromotedLayers(s.renderer.ActiveLayerIDs())
}
