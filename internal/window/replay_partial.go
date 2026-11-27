package window

import (
	"github.com/chinmay-sawant/blinkless/layout"
	"github.com/chinmay-sawant/ownframe/internal/replay"
	"github.com/hajimehoshi/ebiten/v2"
	"image"
)

// drawBackdrop resolves the entire scene into a texture for backdrop shaders.
func (s *shell) drawBackdrop(dst *ebiten.Image, taker dirtyTaker, sameSurface bool) {
	display := s.display
	generation := s.app.Generation()
	scroll := image.Pt(s.scrollX, s.scrollY)
	bounds := dst.Bounds()
	dirty := false
	if taker != nil {
		rects, available := taker.TakeDirtyRects()
		dirty = available && len(rects) > 0
	}
	sameBuf := s.scaleBuf != nil && s.scaleBuf.Bounds().Size() == bounds.Size()
	clean := sameSurface && taker != nil && !s.fullReplay() && !dirty && sameBuf &&
		s.partial.gen == generation && s.partial.scroll == scroll &&
		s.partial.backdropDisplay == display && !s.frameOptIn
	frame := s.ensureScaleBuf(bounds.Dx(), bounds.Dy())
	if !clean {
		s.paintFullReplay(frame, display)
	}
	op := &ebiten.DrawImageOptions{Blend: ebiten.BlendCopy}
	op.GeoM.Translate(float64(bounds.Min.X), float64(bounds.Min.Y))
	dst.DrawImage(frame, op)
	s.partial.gen = generation
	s.partial.scroll = scroll
	s.partial.mode = repaintBlit
	s.partial.whole = false
	s.partial.backdropClean = clean
	s.partial.backdropDisplay = display
}

// drawTileSurface repaints damaged tiles and presents the affected area.
func (s *shell) drawTileSurface(dst *ebiten.Image, frame surfaceFrame, stale bool) {
	display := s.display
	contentW, contentH := frame.content.X, frame.content.Y
	scroll := image.Pt(s.scrollX, s.scrollY)
	whole := scroll != s.partial.scroll
	s.partial.scroll = scroll

	// Prefer top-composited retained layers (BeginViewportLayers) so moving
	// layers composite onto dst in blitFrame without repainting static tiles
	// every frame. Fall back to in-tile layer replay (BeginLayers) when layers
	// are covered by later static operations and cannot safely composite on top.
	prevTileLayers := s.partial.tileLayers
	specs := s.layerSpecs(display)
	gen := s.layerGeneration()
	var layerDamage []image.Rectangle
	var topLayers bool
	if s.partial.tileFallback && s.partial.tileFallbackGen == gen && len(specs) > 0 {
		layerDamage = s.renderer.BeginLayers(display, gen, specs)
	} else {
		s.partial.tileFallback = false
		layerDamage = s.renderer.BeginViewportLayers(display, gen, specs, s.viewportPinZ())
		topLayers = s.renderer.LayersActive(display)
		if !topLayers && len(specs) > 0 {
			var prevTopo []image.Rectangle
			if topo := s.renderer.LayerTopologyDamage(); len(topo) > 0 {
				prevTopo = append([]image.Rectangle(nil), topo...)
			}
			layerDamage = s.renderer.BeginLayers(display, gen, specs)
			if s.renderer.LayersActive(display) {
				s.partial.tileFallback = true
				s.partial.tileFallbackGen = gen
			}
			if len(prevTopo) > 0 {
				layerDamage = append(prevTopo, layerDamage...)
			}
		}
	}
	s.partial.tileLayers = topLayers
	if topLayers {
		s.syncPromotedLayers(display)
		layerDamage = s.renderer.LayerTopologyDamage()
		if !prevTileLayers && len(layerDamage) == 0 && len(specs) > 0 {
			// Switching from in-tile layers to top-composited layers: rebuild
			// the tiles so previously baked layer pixels do not remain underneath.
			layerDamage = []image.Rectangle{image.Rect(0, 0, contentW, contentH)}
		}
	} else {
		s.clearPromotedLayers()
		if prevTileLayers {
			whole = true
		}
	}

	clipX := s.dockClipX(contentW)
	rects, ok := frame.damage.TakeDirtyRects()
	if len(layerDamage) > 0 {
		content := image.Rect(0, 0, contentW, contentH)
		combined := make([]image.Rectangle, 0, len(rects)+len(layerDamage))
		if ok {
			combined = append(combined, rects...)
		}
		for _, rect := range layerDamage {
			if inside := rect.Intersect(content); !inside.Empty() {
				combined = append(combined, inside)
			}
		}
		if len(combined) > 0 {
			rects = combined
			ok = true
		}
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
type layerBoxPromoter interface {
	SetPromotedLayerBoxes([]int)
}

func (s *shell) syncPromotedLayers(d *layout.Display) {
	if promoter, ok := s.app.(layerPromoter); ok {
		promoter.SetPromotedLayers(s.renderer.ActiveLayerIDs())
	}
	if boxPromoter, ok := s.app.(layerBoxPromoter); ok {
		boxPromoter.SetPromotedLayerBoxes(s.renderer.ActiveLayerBoxIndices(d))
	}
}

func (s *shell) clearPromotedLayers() {
	if promoter, ok := s.app.(layerPromoter); ok {
		promoter.SetPromotedLayers(nil)
	}
	if boxPromoter, ok := s.app.(layerBoxPromoter); ok {
		boxPromoter.SetPromotedLayerBoxes(nil)
	}
}

// demoteLayers drops every retained layer and clears the page's promoted set
// when leaving an active layer-compositing path.
func (s *shell) demoteLayers() {
	s.partial.tileLayers = false
	s.partial.tileFallback = false
	s.renderer.DemoteLayers()
	s.clearPromotedLayers()
}

// ticking reports a page with a frame callback. A Page marks mutated ops
// itself and stays on the buffer path.
type ticking interface {
	Ticking() bool
}

// partialState is the frame's retained pixels and the work that keeps them
// current. tiles is the content as a grid of replay tiles: a dirty rect
// repaints only the tiles it covers and the frame blits only damaged tiles.
// buf is the grid's retained-content handle (replay_tile_range.go).
//
// last is what the previous frame repainted, the grid's full rect on a
// rebuild or a blit and one rect per plan otherwise. plans, chrome, and whole
// are this frame's screen blits, kept for a reader that wants them; the
// tiled frame performs them itself. whole copies the content at scroll, the
// offset the last frame was blitted at. focus and screen are the last replay
// frame's window focus and screen size, so a focus regain or a resize
// repaints in full. paints counts the tiles the
// frame repainted and paintPx their area; blits and blitPx count the tile
// blits; pins counts post-blit chrome draws. clipX is the dock clip.
type partialState struct {
	buf             *ebiten.Image
	tiles           *tileGrid
	gen             uint64
	last            []image.Rectangle
	plans           []repaintPlan
	chrome          []image.Rectangle
	blitRects       []image.Rectangle
	mode            repaintMode
	backdropClean   bool
	backdropDisplay *layout.Display
	paints          int
	paintPx         int
	blits           int
	blitPx          int
	pins            int
	focus           bool
	whole           bool
	tileLayers      bool
	tileFallback    bool
	tileFallbackGen uint64
	scroll          image.Point
	screen          image.Point
	clipX           int
	blitOpts        ebiten.DrawImageOptions
}

// gridCurrent reports that the tile grid covers exactly this content size.
func (st partialState) gridCurrent(width, height int) bool {
	g := st.tiles

	return g != nil && g.width == width && g.height == height
}
