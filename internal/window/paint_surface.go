package window

import (
	"image"

	"github.com/chinmay-sawant/blinkless/layout"
	"github.com/hajimehoshi/ebiten/v2"
)

type paintSurface uint8

const (
	surfaceUnset paintSurface = iota
	surfaceFitted
	surfaceBitmapView
	surfaceBitmap
	surfaceSticky
	surfaceScaled
	surfaceDirect
	surfaceBackdrop
	surfaceViewport
	surfaceTiles
)

type surfaceFrame struct {
	mode    paintSurface
	damage  dirtyTaker
	content image.Point
}

// drawContent selects a surface once. Both its resources and subsequent
// idle decision follow the mode that actually painted the frame.
func (s *shell) drawContent(dst *ebiten.Image) {
	var frame surfaceFrame
	switch {
	case s.viewLocked():
		frame.mode = surfaceFitted
	case s.bitmapView.img != nil:
		frame.mode = surfaceBitmapView
	case s.display == nil:
		frame.mode = surfaceBitmap
	case s.viewportSticky:
		frame.mode = surfaceSticky
	case s.stretched() || s.zoom() != 1:
		frame.mode = surfaceScaled
	default:
		frame = s.replaySurface()
	}
	s.paintFrame(dst, frame)
}

func (s *shell) replaySurface() surfaceFrame {
	damage, canTake := s.damageTaker()
	if s.renderer.HasBackdropFilter(s.display) {
		return surfaceFrame{mode: surfaceBackdrop, damage: damage}
	}
	if !canTake || s.fullReplay() || !s.bufferOrderSafe() {
		return surfaceFrame{mode: surfaceDirect}
	}
	w, h := s.contentSize()
	if oversized(w, h) || s.gridTooBig(w, h) || s.windowing() {
		return surfaceFrame{mode: surfaceViewport, damage: damage}
	}
	return surfaceFrame{mode: surfaceTiles, damage: damage, content: image.Pt(w, h)}
}

// usePaintSurface is the transition for the replay buffers and promoted
// layers. Transient scaling textures are released before drawing again.
// Backdrop replay can reuse its texture only across consecutive filtered
// frames. A direct replay can hold the old grid until the next mode switch.
func (s *shell) usePaintSurface(next paintSurface) bool {
	previous := s.surface
	s.surface = next

	hadViewport := s.viewport.buf != nil || s.viewport.valid
	if next != surfaceViewport {
		s.disposeViewport()
	}
	if next != surfaceBackdrop || previous != surfaceBackdrop {
		s.dropScaleBuf()
	}
	if next != surfaceBackdrop {
		s.partial.backdropDisplay = nil
		s.partial.backdropClean = false
	}

	switch next {
	case surfaceTiles:
		s.dropRetainedTiles()
		if previous == surfaceDirect {
			// Direct replay may have demoted layers whose pixels the old
			// grid did not paint. Rebuild the grid before presenting it again.
			s.dropTiles()
		}
	case surfaceViewport:
		if s.partial.tiles != nil {
			s.dropTiles()
			s.partial = partialState{}
		}
	case surfaceDirect:
		s.dropRetainedTiles()
	default:
		s.dropTiles()
	}
	if next != surfaceViewport && (next != surfaceTiles || hadViewport) {
		s.demoteLayers()
	}

	s.partial.paints, s.partial.paintPx, s.partial.pins = 0, 0, 0
	s.partial.blits, s.partial.blitPx = 0, 0
	return previous == next
}

func (s *shell) paintFrame(dst *ebiten.Image, frame surfaceFrame) {
	display := s.display
	if frame.mode == surfaceSticky {
		display = s.stickyDisplay()
		if display != s.display {
			defer s.renderer.EvictDisplay(display)
		}
	}
	sameSurface := s.usePaintSurface(frame.mode)
	switch frame.mode {
	case surfaceBackdrop:
		s.drawBackdrop(dst, frame.damage, sameSurface)
		return
	case surfaceViewport:
		stale := s.frameStale(dst) || !sameSurface
		s.drawViewport(dst, s.display, frame.damage)
		if stale {
			s.viewport.blitOnly = false
		}
		return
	case surfaceTiles:
		s.drawTileSurface(dst, frame, s.frameStale(dst) || !sameSurface)
		return
	}

	dst.Fill(s.pageBackground())
	switch frame.mode {
	case surfaceFitted:
		s.drawFitted(dst)
	case surfaceBitmapView:
		dst.DrawImage(s.bitmapView.img, nil)
	case surfaceBitmap:
		s.drawBitmapPage(dst, s.zoom())
	case surfaceSticky:
		if s.zoom() == 1 && !s.stretched() {
			s.directReplay(dst, display)
		} else {
			s.drawReplayScaled(dst, display)
		}
	case surfaceScaled:
		s.drawReplayScaled(dst, s.display)
	case surfaceDirect:
		s.directReplay(dst, s.display)
	}
}

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

	buf := s.ensureScaleBuf(w, h)

	fillBackground(buf, s.pageBackground())
	paint(buf)

	var op ebiten.DrawImageOptions
	op.Filter = ebiten.FilterLinear
	op.GeoM.Scale(scale, scale)
	op.GeoM.Translate(ox, oy)
	dst.DrawImage(buf, &op)
}

// stickyDisplay projects the scroll-dependent operations without changing
// the source display stored on the shell.
func (s *shell) stickyDisplay() *layout.Display {
	s.syncStickyViewport()
	width, height := s.viewSize()
	zoom := s.zoom()
	return layout.StickyView(
		s.display,
		float64(s.scrollX)/zoom,
		float64(s.scrollY)/zoom,
		float64(width)/zoom,
		float64(height)/zoom,
	)
}
