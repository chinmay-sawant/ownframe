package window

import (
	"image"

	"github.com/chinmay-sawant/blinkless/layout"
	"github.com/chinmay-sawant/ownframe/internal/replay"
	"github.com/hajimehoshi/ebiten/v2"
)

type viewportKey struct {
	generation uint64
	bounds     image.Rectangle
}

type viewportState struct {
	buf   *ebiten.Image
	key   viewportKey
	area  image.Rectangle
	valid bool
}

func (s *shell) drawViewport(dst *ebiten.Image, d *layout.Display, taker dirtyTaker) {
	s.dropContentBuffers()
	bounds := dst.Bounds()
	if bounds.Empty() {
		return
	}
	key := viewportKey{generation: s.seq, bounds: bounds}
	// One cached image, one clipped replay: a single dirty rect is used as
	// is, several use their bounding box. viewportDirtyBounds has the why.
	rects, changed := taker.TakeDirtyRects()
	dirty := viewportDirtyBounds(rects)
	wanted := bounds.Add(image.Pt(s.scrollX, s.scrollY))
	full := !s.viewport.valid || s.viewport.key.bounds != bounds || !wanted.In(s.viewport.area)
	full = full || (s.viewport.key.generation != key.generation && !changed)
	if full {
		s.setViewportArea(bounds)
		s.paintViewport(d, s.viewport.area, s.viewportPinZ())
	} else if changed {
		s.paintViewport(d, dirty, s.viewportPinZ())
	}
	s.viewport.key, s.viewport.valid = key, true
	var options ebiten.DrawImageOptions
	options.GeoM.Translate(float64(s.viewport.area.Min.X-s.scrollX), float64(s.viewport.area.Min.Y-s.scrollY))
	dst.DrawImage(s.viewport.buf, &options)
	replay.DrawFixed(dst, d)
	replay.DrawPinnedZ(dst, d, s.viewportPinZ(), -float64(s.redrawX), -float64(s.redrawY))
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
