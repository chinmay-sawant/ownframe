package window

import (
	"github.com/hajimehoshi/ebiten/v2"
	"image"

	"github.com/chinmay-sawant/blinkless/layout"
)

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

	if s.replayBuf == nil || s.replayBuf.Bounds().Dx() != w || s.replayBuf.Bounds().Dy() != h {
		if s.replayBuf != nil {
			s.replayBuf.Deallocate()
		}

		s.replayBuf = ebiten.NewImageWithOptions(image.Rect(0, 0, w, h), &ebiten.NewImageOptions{Unmanaged: true})
	}

	s.replayBuf.Clear()
	s.replayBuf.Fill(s.pageBackground())
	s.renderer.DrawVisible(s.replayBuf, s.display, 0, 0)

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

	dst.DrawImage(s.replayBuf, &op)
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
