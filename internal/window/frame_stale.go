package window

import "github.com/hajimehoshi/ebiten/v2"

// frameStale reports a frame whose retained pixels cannot be trusted: the
// first frame, a focus regain after the OS may have dropped the surface, or a
// screen size change, or a host that clears its surface every frame. It records the focus and size it saw, so the next frame
// compares against them. Ebiten can still clear after an occlusion; a
// wrong-looking window right after a minimize is the trade.
func (s *shell) frameStale(screen *ebiten.Image) bool {
	focus := ebiten.IsFocused()
	size := screen.Bounds().Size()
	stale := ebiten.IsScreenClearedEveryFrame() || !s.frameSeen() || s.partial.screen != size || focus && !s.partial.focus
	s.partial.focus, s.partial.screen = focus, size

	return stale
}
