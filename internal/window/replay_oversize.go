package window

import (
	"github.com/hajimehoshi/ebiten/v2"

	"github.com/chinmay-sawant/ownframe/internal/host"
)

// maxGridViewports bounds the content-sized tile grid against the window.
// The grid holds every content pixel, so it may cost this many viewport
// buffers before the window takes the viewport path, which keeps only a
// window-sized buffer.
const maxGridViewports = 4

// oversized reports a content size no Ebiten image can hold. A zero device
// limit means the game is not running, so nothing is oversized. A large but
// holdable size keeps its scale: only the tile grid has a budget.
func oversized(w, h int) bool {
	m := ebiten.MaxImageSize()
	if m <= 0 {
		return false
	}

	return w > m || h > m
}

// gridTooBig reports that the content-sized tile grid would cost far more
// memory than the window keeps. The measure is a ratio of the content area
// to the window area, not a pixel count: a full-screen page on a 4K display
// is large in pixels, but its grid is no bigger than the viewport buffer the
// window keeps either way. Only a page several times the window takes the
// viewport path, which bounds its memory to the window.
func (s *shell) gridTooBig(contentW, contentH int) bool {
	winW, winH := s.screenW, s.screenH
	if contentW <= 0 || contentH <= 0 || winW <= 0 || winH <= 0 {
		return false
	}

	return int64(contentW)*int64(contentH) > int64(maxGridViewports)*int64(winW)*int64(winH)
}

// isTicking reports a screen with a frame callback. A Page marks mutated
// ops itself and stays on the buffer path; a ticker without FrameDirty
// still replays in full.
func (s *shell) isTicking() bool {
	t, ok := s.app.(ticking)

	return ok && t.Ticking()
}

// viewportPinZ is the z layer of the page's viewport-pinned operations.
func (s *shell) viewportPinZ() int {
	if pinner, ok := s.app.(host.ViewportPinner); ok {
		return pinner.ViewportPinZ()
	}
	return 0
}
