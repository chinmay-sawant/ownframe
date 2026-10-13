package window

import (
	"github.com/hajimehoshi/ebiten/v2"

	"github.com/chinmay-sawant/blinkless/layout"
)

// devRefresh samples the page stats and drops a box a new generation no
// longer has. It runs after the frame's picture is synced.
func (s *shell) devRefresh() {
	insp := s.devInspector()
	if insp == nil || !s.dev.on {
		return
	}

	if sink, ok := s.app.(drawTimer); ok {
		sink.SetDrawTime(s.dev.draw)
	}

	s.dev.stats = insp.Stats()

	gen := s.app.Generation()
	if gen == s.dev.gen {
		return
	}

	s.dev.gen = gen
	s.dev.haveOp = false
	boxes := s.app.Boxes()

	if s.dev.haveHov && !devHasBox(boxes, s.dev.hovered) {
		s.dev.haveHov = false
	}

	if s.dev.havePin && !devHasBox(boxes, s.dev.pinned) {
		s.dev.havePin = false
	}
}

// devClose clears the picks and sends one hover for the current cursor, so a
// stale :hover does not stick after the overlay closes.
func (s *shell) devClose() error {
	s.dev.haveHov = false
	s.dev.havePin = false

	x, y := ebiten.CursorPosition()
	frameW, frameH := s.frameSize()
	viewW, viewH := s.viewSize()
	px, py := contentPoint(x, y, s.scrollX, s.scrollY, s.stretched(), frameW, frameH, viewW, viewH)

	return s.app.Hover(s.ctx, px, py)
}

// devHasBox reports whether boxes still holds the same element. An id is
// enough when it has one; a box without an id matches its own rectangle.
func devHasBox(boxes []layout.Box, want layout.Box) bool {
	for _, b := range boxes {
		if want.ID != "" {
			if b.ID == want.ID {
				return true
			}

			continue
		}

		if b.Tag == want.Tag && b.X == want.X && b.Y == want.Y && b.W == want.W && b.H == want.H {
			return true
		}
	}

	return false
}
