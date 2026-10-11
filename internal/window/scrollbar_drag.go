package window

const (
	dragNone = iota
	dragVertical
	dragHorizontal
)

// pointerScrollbar handles presses and drags on a scrollbar thumb. It reports
// whether the event was consumed, so content clicks are skipped.
func (s *shell) pointerScrollbar(x, y int, down, pressed bool) bool {
	if !s.scrollbarsEnabled() {
		s.dragAxis = dragNone
		return false
	}

	if !down || s.dragAxis == dragNone {
		if axis, grab, ok := s.scrollbarGrab(x, y); ok {
			s.applyThumbCursor(axis)

			if pressed {
				s.dragAxis = axis
				s.dragGrab = grab

				return true
			}
		}
	}

	if !down {
		s.dragAxis = dragNone

		return false
	}

	if s.dragAxis == dragNone {
		return false
	}

	s.dragScrollbar(x, y)

	return true
}

// scrollbarGrab reports the drag axis and the cursor offset inside the thumb
// when the point sits on a scrollbar thumb.
func (s *shell) scrollbarGrab(x, y int) (int, float64, bool) {
	if !s.scrollbarsEnabled() {
		return dragNone, 0, false
	}

	contentW, contentH := s.contentSize()

	if scrollbarVisible(contentH, s.screenH) && x >= s.screenW-scrollbarThickness {
		pos, length := scrollbarThumb(s.screenH, contentH, s.screenH, s.scrollY)
		if float64(y) >= float64(pos) && float64(y) <= float64(pos+length) {
			return dragVertical, float64(y) - float64(pos), true
		}
	}

	if scrollbarVisible(contentW, s.screenW) && y >= s.screenH-scrollbarThickness {
		pos, length := scrollbarThumb(s.screenW, contentW, s.screenW, s.scrollX)
		if float64(x) >= float64(pos) && float64(x) <= float64(pos+length) {
			return dragHorizontal, float64(x) - float64(pos), true
		}
	}

	return dragNone, 0, false
}
