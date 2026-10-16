package window

import "time"

// touchMove applies one frame of finger movement. A claimed long press
// extends the selection and never scrolls. Any other move scrolls the page,
// clamped to the content ends.
func (s *shell) touchMove(u touchUpdate, now []touchPos, frameW, frameH int) error {
	if u.dx == 0 && u.dy == 0 {
		return nil
	}

	if s.hold.claimed && len(now) == 1 {
		px, py := s.contentAt(now[0].x, now[0].y, frameW, frameH)

		return s.dragAt(px, py)
	}

	s.hold.cancel()
	if err := s.ptr().Release(s.ctx); err != nil {
		return err
	}

	if s.stretched() || !s.allowPageScroll() {
		return nil
	}

	dx, dy := s.scrollMotion.move(u.dx, u.dy, time.Now())
	contentW, contentH := s.contentSize()
	viewW, viewH := s.viewSize()
	s.scrollX, s.scrollY = clampScroll(
		s.scrollX+dx, s.scrollY+dy, contentW, contentH, viewW, viewH,
	)

	return nil
}
