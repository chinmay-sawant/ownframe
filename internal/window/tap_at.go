package window

// tapAt forwards a tap as press, caret, click, release. The caret lands
// before the click handler runs, matching pressAt, and a page without the
// selector keeps the old press, click, release behavior. The tap ends the
// press, so the held-press watch is disarmed first: a touch tap the window
// delivered itself must not fire a long press after the finger is up.
func (s *shell) tapAt(px, py float64) error {
	s.hold.release()

	if err := s.app.Press(s.ctx, px, py); err != nil {
		return err
	}

	return s.finishTap(px, py)
}

func (s *shell) finishTap(px, py float64) error {
	s.hold.release()
	if sel, ok := s.app.(selector); ok {
		if err := sel.SelectAt(s.ctx, px, py); err != nil {
			return err
		}
	}

	if err := s.app.Click(s.ctx, px, py); err != nil {
		return err
	}

	return s.app.Release(s.ctx)
}
