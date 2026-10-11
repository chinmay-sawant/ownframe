package window

// hoverCursor sends page hover, then picks the cursor. Over a scrollbar
// thumb the resize shape pointerScrollbar applied stands, so the page
// shape never overwrites it.
func (s *shell) hoverCursor(x, y int, px, py float64) error {
	if err := s.app.Hover(s.ctx, px, py); err != nil {
		return err
	}

	s.applyCursor()

	return nil
}
