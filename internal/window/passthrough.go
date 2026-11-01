package window

import "github.com/hajimehoshi/ebiten/v2"

func (s *shell) updatePassthrough() {
	if s.interactive == nil {
		return
	}
	x, y := s.cursorPosition()
	pass := !s.windowDrag.active && !s.interactive(x, y)
	if pass != s.passthrough {
		ebiten.SetWindowMousePassthrough(pass)
		s.passthrough = pass
	}
}
