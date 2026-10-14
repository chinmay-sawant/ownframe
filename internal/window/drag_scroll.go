package window

// edgeScroll returns one scroll step when a drag point sits past the top or
// bottom edge of the viewport.
func edgeScroll(y, viewH, step int) int {
	if y < 0 {
		return -step
	}

	if y >= viewH {
		return step
	}

	return 0
}

// dragScroll scrolls while a selection drag sits past an edge. It reuses the
// wheel clamp, so the offset stops at the content ends.
func (s *shell) dragScroll(y int) {
	if !s.dragActive || s.stretched() || !s.allowPageScroll() {
		return
	}

	_, viewH := s.viewSize()

	step := edgeScroll(y, viewH, scrollStep)
	if step == 0 {
		return
	}

	_, contentH := s.contentSize()
	s.scrollY = clampAxis(s.scrollY+step, contentH-viewH)
}
