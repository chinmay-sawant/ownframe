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
	pressed := pressedNow(down, s.mouseDown)
	s.mouseDown = down
	released := inpututil.IsMouseButtonJustReleased(ebiten.MouseButtonLeft)

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
		return s.devPick(px, py, ebiten.IsKeyPressed(ebiten.KeyAlt))
	}

	if released && s.dev.forward {
		s.dev.forward = false

		return s.app.Release(s.ctx)
	}

	return nil
}

// devPick handles a press inside the content while the overlay is on. Alt
// forwards press, caret, and click to the page; picking is only active in
// picker mode, otherwise the press acts on the page control under the cursor.
func (s *shell) devPick(px, py float64, alt bool) error {
	if alt || !s.devPickerOn() {
		s.dev.forward = true

		if err := s.app.Press(s.ctx, px, py); err != nil {
			return err
		}

		if sel, ok := s.app.(selector); ok {
			if err := sel.SelectAt(s.ctx, px, py); err != nil {
				return err
			}
		}

		return s.app.Click(s.ctx, px, py)
	}

	box, ok := devBoxAt(s.app.Boxes(), px, py)
	if !ok {
		s.devDropPin()

		return nil
	}

	if s.dev.havePin && devSameBox(s.dev.pinned, box) {
		s.devDropPin()

		return nil
	}

	s.dev.pinned = box
	s.dev.havePin = true
	s.devShowBoxOps(box, px, py)

	return nil
}
