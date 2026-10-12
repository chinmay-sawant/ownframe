package window

import (
	"github.com/hajimehoshi/ebiten/v2"

	"github.com/chinmay-sawant/blinkless/layout"

	"github.com/chinmay-sawant/ownframe/internal/replay"
)

// blitWhole fills the screen with the page background and shows the whole
// buffer. The buffer is current for the frame, so this replaces a second full
// replay of the display list on frames that only need the screen repainted.
func (s *shell) blitWhole(dst *ebiten.Image) {
	dst.Fill(s.pageBackground())
	s.blitBuffer(dst, nil, nil, true)
}

// applyRepaint brings the persistent buffer up to date for this frame.
func (s *shell) applyRepaint(display *layout.Display, plan repaintPlan) {
	switch plan.mode {
	case repaintBuffer:
		s.ensureBuffer(plan.rect.Dx(), plan.rect.Dy())
		if s.partial.buf == nil {
			return
		}

		s.partial.buf.Fill(s.pageBackground())
		replay.DrawUnfixedZ(s.partial.buf, display, s.viewportPinZ())
	case repaintRect:
		if s.partial.buf == nil {
			return
		}

		if sub, ok := s.partial.buf.SubImage(plan.rect).(*ebiten.Image); ok {
			sub.Fill(s.pageBackground())
		}

		replay.DrawRectUnfixedZ(s.partial.buf, display, plan.rect, s.viewportPinZ())
	}

	s.partial.gen = s.app.Generation()
}

// paintFullReplay paints the whole screen: the page background, then the
// display list. It is the full branch of the replay path, where the buffer
// alone cannot supply the frame.
func (s *shell) paintFullReplay(dst *ebiten.Image, display *layout.Display) {
	dst.Fill(s.pageBackground())
	s.directReplay(dst, display)
}

// ensureBuffer keeps one content-sized image for the replay path. The content
// covers the canvas and any box that overflows it, so a scrolled page stays
// painted. The image is rebuilt when the content size changes.
func (s *shell) ensureBuffer(width, height int) {
	if s.partial.buf != nil && s.partial.width == width && s.partial.height == height {
		return
	}

	if s.partial.buf != nil {
		s.partial.buf.Dispose()
		s.partial.buf = nil
	}

	if width <= 0 || height <= 0 {
		return
	}

	s.partial.buf = ebiten.NewImage(width, height)
	s.partial.width, s.partial.height = width, height
}
