package window

// refreshState re-resolves hover and press against the boxes the committed
// relayout just produced. The point is the cursor from the last pointer
// pass, mapped into the new frame. Hover and Press redraw only when the id
// under it changed, so a relayout that moves no control costs no cascade.
func (s *shell) refreshState() error {
	px, py := s.contentPointAt(s.cursorX, s.cursorY)

	if err := s.ptr().Hover(s.ctx, px, py); err != nil {
		return err
	}

	if !s.mouseDown {
		return nil
	}

	return s.ptr().Press(s.ctx, px, py)
}
