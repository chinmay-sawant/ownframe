package window

import (
	"github.com/hajimehoshi/ebiten/v2"

	"github.com/chinmay-sawant/blinkless/layout"
	"github.com/chinmay-sawant/ownframe/internal/host"
	"github.com/chinmay-sawant/ownframe/internal/replay"
)

// maxTilePixels bounds the content-sized tile grid. Past it the viewport
// buffer is smaller than the full tile grid, so the window uses that path.
const maxTilePixels = 4 << 20

// oversized reports a content size no Ebiten image can hold, or one whose
// tile grid would retain more pixels than the viewport budget. A zero device
// limit means the game is not running, so only the pixel budget applies.
func oversized(w, h int) bool {
	if w > 0 && h > 0 && w*h > maxTilePixels {
		return true
	}

	m := ebiten.MaxImageSize()
	if m <= 0 {
		return false
	}

	return w > m || h > m
}

// isTicking reports a screen that changes operations from a frame callback.
// A tick changes operations in place without a dirty rect, so such a page
// keeps the full replay.
func (s *shell) isTicking() bool {
	t, ok := s.app.(ticking)

	return ok && t.Ticking()
}

// directReplay draws the list straight to the screen. Ticking pages use it
// because they have no dirty rect; oversized content uses it because no
// buffer can hold it. Viewport-pinned layers draw at the offset the display
// was built with, so they stay at the viewport while the page scrolls.
func (s *shell) directReplay(dst *ebiten.Image, display *layout.Display) {
	replay.DrawVisiblePinned(dst, display,
		-float64(s.scrollX), -float64(s.scrollY),
		s.viewportPinZ(), -float64(s.redrawX), -float64(s.redrawY))
}

func (s *shell) viewportPinZ() int {
	if pinner, ok := s.app.(host.ViewportPinner); ok {
		return pinner.ViewportPinZ()
	}
	return 0
}
