package window

import (
	"github.com/hajimehoshi/ebiten/v2"
	"image"

	"github.com/chinmay-sawant/blinkless/layout"
)

// ensureScaleBuf keeps the one shared content-sized buffer both scaled draws
// reuse: the zoomed replay and the fitted replay never run on the same frame,
// so they share one image instead of pinning two content-sized buffers.
func (s *shell) ensureScaleBuf(w, h int) *ebiten.Image {
	if s.scaleBuf == nil || s.scaleBuf.Bounds().Dx() != w || s.scaleBuf.Bounds().Dy() != h {
		if s.scaleBuf != nil {
			s.scaleBuf.Deallocate()
		}

		// This image alternates between a render target and a blit source.
		// Keep it out of the sprite atlas so those switches never migrate it.
		s.scaleBuf = ebiten.NewImageWithOptions(image.Rect(0, 0, w, h), &ebiten.NewImageOptions{Unmanaged: true})
	}

	return s.scaleBuf
}

// dropScaleBuf releases the shared scaled buffer. Tile, viewport, and direct
// frames never read it, so they drop it instead of pinning a stale copy.
func (s *shell) dropScaleBuf() {
	if s.scaleBuf != nil {
		s.scaleBuf.Deallocate()
		s.scaleBuf = nil
	}
}

// drawReplayScaled draws the display list through an offscreen buffer and
// scales it to the window, matching the bitmap path on a resized or clamped
// window.
func (s *shell) drawReplayScaled(dst *ebiten.Image) {
	w, h := s.display.Width, s.display.Height
	viewW, viewH := s.viewSize()
	if w <= 0 || h <= 0 || viewW <= 0 || viewH <= 0 {
		return
	}

	if oversized(w, h) {
		// No image can hold the page, so the draw drops the scale.
		s.directReplay(dst, s.display)

		return
	}

	// The scaled frame reads neither the tile grid nor the viewport cache.
	s.dropDisplayBuffers()
	buf := s.ensureScaleBuf(w, h)

	buf.Clear()
	buf.Fill(s.pageBackground())
	s.renderer.DrawVisible(buf, s.display, 0, 0)

	var op ebiten.DrawImageOptions
	op.Filter = ebiten.FilterLinear

	if s.stretched() {
		op.GeoM.Scale(
			float64(viewW)/float64(w)*s.zoom(),
			float64(viewH)/float64(h)*s.zoom(),
		)
	} else {
		op.GeoM.Scale(s.zoom(), s.zoom())
		op.GeoM.Translate(-float64(s.scrollX), -float64(s.scrollY))
	}

	dst.DrawImage(buf, &op)
}

// directReplay draws the list straight to the screen. Ticking pages use it
// because they have no dirty rect; oversized content uses it because no
// buffer can hold it. Viewport-pinned layers draw at the offset the display
// was built with, so they stay at the viewport while the page scrolls.
func (s *shell) directReplay(dst *ebiten.Image, display *layout.Display) {
	s.renderer.DrawVisiblePinned(dst, display,
		-float64(s.scrollX), -float64(s.scrollY),
		s.viewportPinZ(), -float64(s.redrawX), -float64(s.redrawY))
}
