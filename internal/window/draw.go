package window

import (
	"time"

	"github.com/hajimehoshi/ebiten/v2"
)

func (s *shell) Draw(screen *ebiten.Image) {
	s.devFrameTime()
	s.perfFrameSample()

	screen.Fill(s.pageBackground())

	if s.fallback {
		defer s.drawBadge(screen)
	}

	start := time.Now()
	s.drawContent(screen)
	s.dev.draw = time.Since(start)
	if s.perf {
		s.dev.stages.draw = s.dev.draw
	}

	tail := time.Now()
	s.drawScrollbars(screen)
	s.drawMenu(screen)

	if s.dev.on {
		s.drawDevTools(screen)
	}
	if s.perf {
		s.dev.stages.present = time.Since(tail)
	}
}

func (s *shell) drawContent(screen *ebiten.Image) {
	if s.viewLocked() {
		s.drawFitted(screen)
		return
	}
	if s.bitmapView.img != nil {
		dst := screen
		dst.DrawImage(s.bitmapView.img, nil)
		return
	}
	zoom := s.zoom()

	if s.display != nil {
		if s.stretched() || zoom != 1 {
			s.drawReplayScaled(screen)

			return
		}

		s.drawReplayPartial(screen)

		return
	}

	if s.img == nil {
		return
	}

	bounds := s.img.Bounds()
	if bounds.Dx() == 0 || bounds.Dy() == 0 || s.screenW == 0 || s.screenH == 0 {
		return
	}

	var op ebiten.DrawImageOptions
	// A page larger than the window keeps its own pixels and scrolls.
	// A size that has not been laid out yet is scaled until the new picture arrives.
	if !s.stretched() {
		if zoom != 1 {
			op.Filter = ebiten.FilterLinear
			op.GeoM.Scale(zoom, zoom)
		}

		op.GeoM.Translate(-float64(s.scrollX), -float64(s.scrollY))
		screen.DrawImage(s.img, &op)

		return
	}

	op.Filter = ebiten.FilterLinear
	op.GeoM.Scale(
		float64(s.screenW)/float64(bounds.Dx())*zoom,
		float64(s.screenH)/float64(bounds.Dy())*zoom,
	)
	screen.DrawImage(s.img, &op)
}
