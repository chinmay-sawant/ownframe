package window

import "github.com/chinmay-sawant/blinkless/layout"

// devOpInBox returns the display index of the topmost painted operation
// whose centre lies inside box, or -1. The centre rule matches internal/frame
// and tolerates a stroke or an ink bound just past the element edge.
func devOpInBox(display *layout.Display, box layout.Box) int {
	if display == nil {
		return -1
	}

	found := -1

	for _, i := range display.Order {
		if i < 0 || i >= len(display.Ops) {
			continue
		}

		op := &display.Ops[i]
		if _, ok := devOpInk(op.Kind); !ok {
			continue
		}

		r := devOpRect(op, display.PointsPerPixel)
		cx, cy := r.X+r.W/2, r.Y+r.H/2
		if cx >= box.X-0.5 && cx <= box.X+box.W+0.5 &&
			cy >= box.Y-0.5 && cy <= box.Y+box.H+0.5 {
			found = i
		}
	}

	return found
}

// devShowBoxOps opens the Ops tab for a pinned box, selects the operation
// under the click when one sits there and the operation that fills the box
// otherwise, and scrolls its row into view.
func (s *shell) devShowBoxOps(box layout.Box, px, py float64) {
	s.dev.tab = devTabOps

	i := devOpAt(s.display, box, px, py)
	if i < 0 {
		i = devOpInBox(s.display, box)
	}

	if i < 0 {
		return
	}

	s.dev.opPick, s.dev.haveOp, s.dev.opPin = i, true, true
	s.devScrollToOp(i)
}

// devDropPin clears the pinned box and the operation the pick selected.
func (s *shell) devDropPin() {
	s.dev.havePin = false

	if s.dev.opPin {
		s.dev.haveOp = false
		s.dev.opPin = false
	}
}

// devScrollToOp moves the dock content so the row for an operation shows.
func (s *shell) devScrollToOp(i int) {
	c := s.dev.content
	if c.H <= 0 {
		return
	}

	lineH := devLineHeight()
	rows := s.devOpsRows()
	s.dev.rows = rows

	for n, row := range rows {
		if row.act != devActOp || row.arg != i {
			continue
		}

		top := float64(n) * lineH
		if top < s.dev.scroll {
			s.dev.scroll = top
		}

		if top+lineH > s.dev.scroll+c.H {
			s.dev.scroll = top + lineH - c.H
		}

		s.devClampScroll()

		return
	}
}
