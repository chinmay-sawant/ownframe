package window

import "github.com/chinmay-sawant/ownframe/internal/host"

// hoverState remembers the page point of the last hover, the element the
// page painted after it, and the built generation it resolved against, so a
// repeat frame can skip it. The cursor fields remember the point,
// generation, and hover the page cursor was last applied for on the page
// path, so a still frame skips the cursor shape scan too.
type hoverState struct {
	x, y   float64
	seq    uint64
	id     string
	have   bool
	cx, cy float64
	cgen   uint64
	cid    string
	cok    bool
}

// due reports whether the page hover needs a fresh pass. A point that moved
// always needs one. A still point on the same built generation reuses the
// last result. A new generation can move a box under the point, so that
// point resolves again too, unless the page paints a different element than
// the one this window pushed: then the page moved its hover itself, and
// pushing the window's stale point would relayout the page a second time for
// a hover it no longer shows. A screen that does not report its hover keeps
// resolving on every generation, as before.
func (h *hoverState) due(px, py float64, built uint64, pageHover string, reports bool) bool {
	if !h.have || h.x != px || h.y != py {
		return true
	}
	if h.seq == built {
		return false
	}

	return !reports || pageHover == h.id
}

// pageHover reports the element the page paints as hovered, when the screen
// can tell.
func (s *shell) pageHover() (string, bool) {
	hoverer, ok := s.app.(host.Hoverer)
	if !ok {
		return "", false
	}

	return hoverer.HoverID(), true
}

// hoverPushed records the hover the window just sent. The generation is the
// one the push left, not the one it resolved against: a hover that redrew
// the page must not read as a new generation on the next frame, or the same
// point would resolve again for a change it caused itself.
func (s *shell) hoverPushed(px, py float64) hoverState {
	id, _ := s.pageHover()

	return hoverState{x: px, y: py, seq: s.app.Generation(), id: id, have: true}
}

// hoverCursor sends page hover, then picks the cursor. Over a scrollbar
// thumb the resize shape pointerScrollbar applied stands, so the page
// shape never overwrites it. The first update runs before the first draw,
// so its hover push is deferred until one frame is committed: resolving it
// redraws the whole page for a cursor the user has not seen a frame under
// yet. Skipping the state record keeps it due, so the next update applies
// it after the first visible frame. A still point on the same generation
// with the same page hover keeps the applied page cursor: re-resolving it
// would walk the boxes and the link operations for the shape the window
// already shows. A thumb or drag path clears that record, so leaving the
// thumb re-applies the page shape.
func (s *shell) hoverCursor(x, y int, px, py float64) error {
	pageHover, reports := s.pageHover()
	pushed := false
	if s.hover.due(px, py, s.seq, pageHover, reports) && s.frameSeen() {
		if err := s.ptr().Hover(s.ctx, px, py); err != nil {
			return err
		}

		s.hover = s.hoverPushed(px, py)
		pushed = true
		pageHover = s.hover.id
	}

	if s.dragAxis != dragNone {
		s.hover.cok = false
		s.applyThumbCursor(s.dragAxis)
		return nil
	}
	if axis, _, ok := s.scrollbarGrab(x, y); ok {
		s.hover.cok = false
		s.applyThumbCursor(axis)
		return nil
	}
	if !pushed && s.hover.cok && s.hover.cx == px && s.hover.cy == py &&
		s.hover.cgen == s.app.Generation() && s.hover.cid == pageHover {
		return nil
	}
	s.applyCursor()
	s.hover.cx, s.hover.cy = px, py
	s.hover.cgen = s.app.Generation()
	s.hover.cid = pageHover
	s.hover.cok = true

	return nil
}
