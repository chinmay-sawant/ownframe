package window

import (
	"github.com/hajimehoshi/ebiten/v2"
)

// drawFitted paints the whole page inside the window.
// One scale serves both axes, and the page stays centered.
func (s *shell) drawFitted(dst *ebiten.Image) {
	scale, ox, oy := s.fit()
	w, h := s.contentSize()
	if w < 1 || h < 1 || scale <= 0 {
		return
	}

	if s.display != nil {
		s.blitFitted(dst, w, h, scale, ox, oy, func(buf *ebiten.Image) {
			s.renderer.Draw(buf, s.display, 0, 0)
		})

		return
	}

	if s.img == nil {
		return
	}

	s.blitFitted(dst, w, h, scale, ox, oy, func(buf *ebiten.Image) {
		buf.DrawImage(s.img, nil)
	})
}

func (s *shell) blitFitted(dst *ebiten.Image, w, h int, scale, ox, oy float64, paint func(*ebiten.Image)) {
	if oversized(w, h) {
		// No image can hold the page, so the fit drops the scale.
		if s.display != nil {
			s.directReplay(dst, s.display)
		}

		return
	}

	// The fitted frame reads neither the tile grid nor the viewport cache.
	s.dropDisplayBuffers()
	buf := s.ensureScaleBuf(w, h)

	buf.Clear()
	buf.Fill(s.pageBackground())
	paint(buf)

	var op ebiten.DrawImageOptions
	op.Filter = ebiten.FilterLinear
	op.GeoM.Scale(scale, scale)
	op.GeoM.Translate(ox, oy)
	dst.DrawImage(buf, &op)
}
