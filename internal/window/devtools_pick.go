package window

import (
	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/inpututil"
)

// devPointer routes the pointer while the overlay is on. The dock owns its
// hits and its resize edge; the content pins the box under the cursor.
// Alt+click forwards press, caret, and click to the page.
func (s *shell) devPointer(x, y int, px, py float64) error {
	down := ebiten.IsMouseButtonPressed(ebiten.MouseButtonLeft)
	released := inpututil.IsMouseButtonJustReleased(ebiten.MouseButtonLeft)
	pressed := pressEdge(down, released, s.mouseDown)
	s.mouseDown = down

	if s.dev.resizing {
		if down {
			s.devResize(x)

			return nil
		}

		s.dev.resizing = false
	}

	if pressed && devOnEdge(s.dev.panel, x, y) {
		s.dev.resizing = true

		return nil
	}

	if devInRect(s.dev.panel, float64(x), float64(y)) {
		if pressed {
			s.devPanelPress(x, y)
		}

		return nil
	}

	if s.devPickerOn() {
		s.devHover(px, py)
	}

	if pressed {
		if err := s.devPick(px, py, ebiten.IsKeyPressed(ebiten.KeyAlt)); err != nil {
			return err
		}

		if released && s.dev.forward {
			s.dev.forward = false

			return s.ptr().Release(s.ctx)
		}

		return nil
	}

	if released && s.dev.forward {
		s.dev.forward = false

		return s.ptr().Release(s.ctx)
	}

	return nil
}
