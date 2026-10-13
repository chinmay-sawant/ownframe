package window

import (
	"github.com/hajimehoshi/ebiten/v2"

	"github.com/chinmay-sawant/ownframe/internal/replay"
)

// drawReplayScaled draws the display list through an offscreen buffer and
// scales it to the window, matching the bitmap path when the page size and
// the window size differ (a resized or clamped window).
func (s *shell) drawReplayScaled(dst *ebiten.Image) {
	w, h := s.display.Width, s.display.Height
	viewW, viewH := s.viewSize()
	if w <= 0 || h <= 0 || viewW <= 0 || viewH <= 0 {
		return
	}

	if oversized(w, h) {
		s.directReplay(dst, s.display)

		return
	}

	if s.replayBuf == nil || s.replayBuf.Bounds().Dx() != w || s.replayBuf.Bounds().Dy() != h {
		if s.replayBuf != nil {
			s.replayBuf.Dispose()
		}

		s.replayBuf = ebiten.NewImage(w, h)
	}

	s.replayBuf.Clear()
	s.replayBuf.Fill(s.pageBackground())
	replay.Draw(s.replayBuf, s.display, 0, 0)

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
