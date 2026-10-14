package window

func (s *shell) dragScrollbar(x, y int) {
	contentW, contentH := s.contentSize()
	viewW, viewH := s.viewSize()

	switch s.dragAxis {
	case dragVertical:
		pos := float64(y) - s.dragGrab
		s.scrollY = clampAxis(scrollbarOffset(pos, viewH, contentH, viewH), contentH-viewH)
	case dragHorizontal:
		pos := float64(x) - s.dragGrab
		s.scrollX = clampAxis(scrollbarOffset(pos, viewW, contentW, viewW), contentW-viewW)
	}
}
