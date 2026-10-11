package window

import (
	"image"

	"github.com/hajimehoshi/ebiten/v2"

	"github.com/chinmay-sawant/ownframe/internal/replay"
)

// dirtyTaker is the page side of the partial replay. TakeDirty returns the
// region that changed since the previous call, in CSS pixels, and clears it.
type dirtyTaker interface {
	TakeDirty() (image.Rectangle, bool)
}

// ticking reports a page with a frame callback. A tick changes operations in
// place without a dirty rect, so such a page keeps the full replay.
type ticking interface {
	Ticking() bool
}

// partialState is the persistent replay buffer and the frame it holds.
type partialState struct {
	buf    *ebiten.Image
	gen    uint64
	width  int
	height int
	last   image.Rectangle
	mode   repaintMode
}

// drawReplayPartial draws a display-list page through the persistent buffer
// when the page can report a dirty rect, and replays the whole list when it
// cannot. The screen is cleared each frame, so the visible page is blitted
// from the buffer every frame; only the buffer repaint is partial.
func (s *shell) drawReplayPartial(dst *ebiten.Image) {
	display := s.display
	if display == nil {
		return
	}

	taker, canTake := s.app.(dirtyTaker)
	contentW, contentH := s.contentSize()
	if !canTake || s.fullReplay() || !s.bufferOrderSafe() {
		s.disposeViewport()
		s.directReplay(dst, display)

		return
	}

	if oversized(contentW, contentH) {
		s.drawViewport(dst, display, taker)
		return
	}
	s.disposeViewport()

	rect, ok := taker.TakeDirty()
	plan := planRepaint(contentW, contentH, s.partial, s.app.Generation(), rect, ok)
	s.applyRepaint(display, plan)
	s.partial.last, s.partial.mode = plan.rect, plan.mode

	if s.partial.buf == nil {
		return
	}

	var op ebiten.DrawImageOptions
	op.GeoM.Translate(-float64(s.scrollX), -float64(s.scrollY))
	dst.DrawImage(s.partial.buf, &op)
	replay.DrawFixed(dst, display)
	replay.DrawPinnedZ(dst, display, s.viewportPinZ(), -float64(s.redrawX), -float64(s.redrawY))
}
