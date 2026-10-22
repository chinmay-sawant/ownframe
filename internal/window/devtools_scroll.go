package window

import "github.com/hajimehoshi/ebiten/v2"

// devScrollLines is the lines one wheel notch moves.
const devScrollLines = 3.0

// devWheel routes the wheel to the dock while the pointer is over it. It
// reports whether the dock ate the wheel, so the page below stays put.
func (s *shell) devWheel() bool {
	if !s.dev.on || !devInRect(s.dev.panel, float64(s.cursorX), float64(s.cursorY)) {
		return false
	}

	_, wheelY := ebiten.Wheel()
	s.dev.scroll -= wheelY * devLineHeight() * devScrollLines
	s.devClampScroll()

	return true
}

// devClampScroll keeps the dock content inside its viewport.
func (s *shell) devClampScroll() {
	max := float64(len(s.dev.rows))*devLineHeight() - s.dev.content.H
	if max < 0 {
		max = 0
	}

	if s.dev.scroll > max {
		s.dev.scroll = max
	}

	if s.dev.scroll < 0 {
		s.dev.scroll = 0
	}
}
