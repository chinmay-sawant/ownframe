package window

// hoverState remembers the page point of the last hover and the built
// generation it resolved against, so a repeat frame can skip it.
type hoverState struct {
	x, y float64
	seq  uint64
	have bool
}

// due reports whether the page hover needs a fresh pass. A still point on
// the same built generation reuses the last result; a new generation can
// move a box under the point, so it always resolves again.
func (h *hoverState) due(px, py float64, seq uint64) bool {
	return !h.have || h.x != px || h.y != py || h.seq != seq
}

// hoverCursor sends page hover, then picks the cursor. Over a scrollbar
// thumb the resize shape pointerScrollbar applied stands, so the page
// shape never overwrites it. The first update runs before the first draw,
// so its hover push is deferred until one frame is committed: resolving it
// redraws the whole page for a cursor the user has not seen a frame under
// yet. Skipping the state record keeps it due, so the next update applies
// it after the first visible frame.
func (s *shell) hoverCursor(x, y int, px, py float64) error {
	if s.hover.due(px, py, s.seq) && s.frameSeen() {
		if err := s.ptr().Hover(s.ctx, px, py); err != nil {
			return err
		}

		s.hover = hoverState{x: px, y: py, seq: s.seq, have: true}
	}

	if s.dragAxis != dragNone {
		s.applyThumbCursor(s.dragAxis)
		return nil
	}
	if axis, _, ok := s.scrollbarGrab(x, y); ok {
		s.applyThumbCursor(axis)
		return nil
	}
	s.applyCursor()

	return nil
}
