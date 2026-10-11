package window

import (
	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/inpututil"
)

// pointer sends hover and clicks. A click needs an up-to-down edge the shell
// saw itself. IsMouseButtonJustPressed can repeat for one press when a
// blocking call, such as the file dialog, stalls the event queue.
func (s *shell) pointer() error {
	if handled, err := s.consumeTouchCancel(ebiten.TouchIDs()); handled || err != nil {
		return err
	}
	x, y := ebiten.CursorPosition()
	s.cursorX, s.cursorY = x, y
	if handled, err := s.moveWindow(x, y); handled || err != nil {
		return err
	}

	if s.pointerScrollbar(x, y) {
		return nil
	}

	frameW, frameH := s.frameSize()
	px, py := s.contentAt(x, y, frameW, frameH)

	if s.dev.on {
		return s.devPointer(x, y, px, py)
	}

	if err := s.hoverCursor(x, y, px, py); err != nil {
		return err
	}

	handled, err := s.menuPointer(x, y)
	if err != nil {
		return err
	}

	down := ebiten.IsMouseButtonPressed(ebiten.MouseButtonLeft)
	if !handled && down && s.mouseDown && !s.gesture.active {
		s.dragScroll(y)
	}

	clicked := false
	if !handled {
		if pressedNow(down, s.mouseDown) {
			clicked = true
			s.holdStart(px, py)
			if err := s.pressAt(px, py); err != nil {
				return err
			}
		} else if down {
			s.hold.move(px, py)
			if err := s.dragAt(px, py); err != nil {
				return err
			}
		}
	}
	s.mouseDown = down

	if inpututil.IsMouseButtonJustReleased(ebiten.MouseButtonLeft) {
		if err := s.releaseAt(); err != nil {
			return err
		}
	}

	return s.touch(clicked, frameW, frameH)
}

// contentPointAt maps the last cursor position into the painted page.
func (s *shell) contentPointAt(x, y int) (float64, float64) {
	frameW, frameH := s.frameSize()

	return s.contentAt(x, y, frameW, frameH)
}

// pressedNow reports the up-to-down edge of a pointer level.
func pressedNow(down, wasDown bool) bool {
	return down && !wasDown
}
