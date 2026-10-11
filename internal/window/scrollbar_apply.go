package window

func (s *shell) dragScrollbar(x, y int) {
	contentW, contentH := s.contentSize()
	switch s.dragAxis {
	case dragVertical:
		pos := float64(y) - s.dragGrab
		s.scrollY = clampAxis(scrollbarOffset(pos, s.screenH, contentH, s.screenH), contentH-s.screenH)
	case dragHorizontal:
		pos := float64(x) - s.dragGrab
		s.scrollX = clampAxis(scrollbarOffset(pos, s.screenW, contentW, s.screenW), contentW-s.screenW)
	}
}
