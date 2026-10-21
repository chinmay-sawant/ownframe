package window

func (s *shell) pointerMouse(
	x, y int, px, py float64, down, released, blocked bool,
) (bool, error) {
	pressed := pressEdge(down, released, s.mouseDown)
	handled := blocked || s.pointerScrollbar(x, y, down, pressed)
	if !handled && down && s.mouseDown && !s.gesture.active {
		s.dragScroll(y)
	}

	clicked := false
	if !handled {
		if pressed {
			clicked = true
			s.holdStart(px, py)
			if err := s.pressAt(px, py); err != nil {
				return false, err
			}
		} else if down {
			s.hold.move(px, py)
			if err := s.dragAt(px, py); err != nil {
				return false, err
			}
		}
	}
	s.mouseDown = down
	if released {
		if err := s.releaseAt(); err != nil {
			return clicked, err
		}
	}
	return clicked, nil
}
