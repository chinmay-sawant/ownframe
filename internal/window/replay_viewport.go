package window

import (
	"image"

	"github.com/chinmay-sawant/blinkless/layout"
	"github.com/hajimehoshi/ebiten/v2"
)

type viewportKey struct {
	generation uint64
	bounds     image.Rectangle
}

type viewportState struct {
	buf        *ebiten.Image
	key        viewportKey
	area       image.Rectangle
	scroll     image.Point
	valid      bool
	blitOnly   bool
	prewarmed  bool
	prewarmGen uint64
}

func (s *shell) drawViewport(dst *ebiten.Image, d *layout.Display, taker dirtyTaker) {
	s.dropContentBuffers()
	bounds := dst.Bounds()
	if bounds.Empty() {
		s.viewport.blitOnly = false
		return
	}
	// setViewportArea may dispose the buffer and zero this state.
	had := s.viewport.valid && s.viewport.buf != nil
	prevScroll := s.viewport.scroll
	scroll := image.Pt(s.scrollX, s.scrollY)
	key := viewportKey{generation: s.seq, bounds: bounds}
	// Bake this frame's retained layers before reading dirty state. The
	// background below leaves their operations as holes, and the layers are
	// composited after the blit, so a moving layer never repaints the page
	// under it.
	specs := s.layerSpecs(d)
	s.renderer.BeginLayers(d, s.layerGeneration(), specs)
	s.syncPromotedLayers()
	// One cached image: a single dirty rect is used as is, and a list is
	// painted rect by rect so a moving chip costs its own box instead of the
	// bounding box of every moving chip.
	rects, changed := taker.TakeDirtyRects()
	// A promotion opens a hole in the background and a demotion closes one.
	// Ordinary layer motion needs neither, because the pixels under a layer
	// stay background.
	if topology := s.renderer.LayerTopologyDamage(); len(topology) > 0 {
		rects = append(rects, topology...)
	}
	changed = changed || len(rects) > 0
	wanted := bounds.Add(scroll)
	full := !s.viewport.valid || s.viewport.key.bounds != bounds || !wanted.In(s.viewport.area)
	full = full || (s.viewport.key.generation != key.generation && !changed)
	if full {
		s.setViewportArea(bounds)
		s.paintViewport(d, s.viewport.area, s.viewportPinZ())
	} else if changed && len(rects) > 0 {
		s.paintViewportRects(d, rects)
	}
	noPaint := !full && !changed
	s.viewport.blitOnly = had && noPaint && scroll == prevScroll
	s.viewport.scroll = scroll
	s.viewport.key, s.viewport.valid = key, true
	var options ebiten.DrawImageOptions
	options.GeoM.Translate(
		float64(s.viewport.area.Min.X-s.scrollX),
		float64(s.viewport.area.Min.Y-s.scrollY),
	)
	dst.DrawImage(s.viewport.buf, &options)
	// Layers composite onto dst in page coordinates, not into the buffer, so
	// they take the scroll offset alone. Adding the buffer origin here put
	// every retained layer one overscan band low once the page scrolled.
	s.renderer.DrawLayers(dst, d, -float64(s.scrollX), -float64(s.scrollY))
	s.renderer.DrawFixed(dst, d)
	s.renderer.DrawPinnedZ(dst, d, s.viewportPinZ(), -float64(s.redrawX), -float64(s.redrawY))
	if gen := s.layerGeneration(); !s.viewport.prewarmed || s.viewport.prewarmGen != gen {
		s.viewport.prewarmed = true
		s.viewport.prewarmGen = gen
		s.prewarmViewport(d, bounds)
	}
}

func (s *shell) disposeViewport() {
	if s.viewport.buf != nil {
		s.viewport.buf.Deallocate()
	}
	s.viewport = viewportState{}
}

func (s *shell) dropContentBuffers() {
	if s.partial.tiles != nil {
		s.partial.tiles.dispose()
		s.partial = partialState{}
	}

	if s.replayBuf != nil {
		s.replayBuf.Deallocate()
		s.replayBuf = nil
	}
}
