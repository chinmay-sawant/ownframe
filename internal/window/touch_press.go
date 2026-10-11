package window

func (s *shell) touchPress(u touchUpdate, now []touchPos, clicked bool, w, h int) error {
	if len(now) >= 2 {
		armed := s.hold.armed || s.hold.claimed
		s.hold.cancel()
		s.scrollMotion.cancel()
		if armed {
			return s.app.Release(s.ctx)
		}
		return nil
	}
	if u.start == nil || clicked {
		return nil
	}
	x, y := s.contentAt(u.start.x, u.start.y, w, h)
	s.holdStart(x, y)
	return s.app.Press(s.ctx, x, y)
}
