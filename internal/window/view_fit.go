package window

import (
	"github.com/hajimehoshi/ebiten/v2"

	"github.com/chinmay-sawant/ownframe/internal/replay"
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
			replay.Draw(buf, s.display, 0, 0)
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

	if s.fitBuf == nil || s.fitBuf.Bounds().Dx() != w || s.fitBuf.Bounds().Dy() != h {
		if s.fitBuf != nil {
			s.fitBuf.Dispose()
		}

		s.fitBuf = ebiten.NewImage(w, h)
	}

	s.fitBuf.Clear()
	s.fitBuf.Fill(s.pageBackground())
	paint(s.fitBuf)

	var op ebiten.DrawImageOptions
	op.Filter = ebiten.FilterLinear
	op.GeoM.Scale(scale, scale)
	op.GeoM.Translate(ox, oy)
	dst.DrawImage(s.fitBuf, &op)
}
