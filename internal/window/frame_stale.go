package window

import "github.com/hajimehoshi/ebiten/v2"

// frameStale reports a frame whose retained pixels cannot be trusted: the
// first frame, a focus regain after the OS may have dropped the surface, or a
// screen size change, or a host that clears its surface every frame.
// It records focus and size for the next frame. Retained desktop surfaces
// can still lose pixels after an occlusion without a focus change.
func (s *shell) frameStale(screen *ebiten.Image) bool {
	focus := ebiten.IsFocused()
	size := screen.Bounds().Size()
	stale := ebiten.IsScreenClearedEveryFrame() || !s.frameSeen() || s.partial.screen != size || focus && !s.partial.focus
	s.partial.focus, s.partial.screen = focus, size

	return stale
}
