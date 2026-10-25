package window

import (
	"github.com/chinmay-sawant/blinkless/layout"
	"github.com/chinmay-sawant/ownframe/internal/host"
	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/inpututil"
)

// devPointer routes the pointer while the overlay is on. The dock owns its
// hits and its resize edge; the content pins the box under the cursor.
// Alt+click forwards press, caret, and click to the page.
func (s *shell) devPointer(x, y int, px, py float64) error {
	if err := s.devPointerHover(x, y, px, py); err != nil {
		return err
	}
	down := ebiten.IsMouseButtonPressed(ebiten.MouseButtonLeft)
	released := inpututil.IsMouseButtonJustReleased(ebiten.MouseButtonLeft)
	pressed := pressEdge(down, released, s.mouseDown)
	s.mouseDown = down

	if s.dev.resizing {
		if down {
			s.devResize(x)

			return nil
		}

		s.dev.resizing = false
	}

	if !s.dev.mobile && pressed && devOnEdge(s.dev.panel, x, y) {
		s.dev.resizing = true

		return nil
	}

	if devInRect(s.dev.panel, float64(x), float64(y)) {
		if pressed {
			s.devPanelPress(x, y)
		}

		return nil
	}

	if pressed {
		if err := s.devPick(px, py, ebiten.IsKeyPressed(ebiten.KeyAlt)); err != nil {
			return err
		}

		if released && s.dev.forward {
			s.dev.forward = false

			return s.ptr().Release(s.ctx)
		}

		return nil
	}

	if released && s.dev.forward {
		s.dev.forward = false

		return s.ptr().Release(s.ctx)
	}

	return nil
}

// devPointerHover keeps normal page hover and cursor feedback unless the
// pointer belongs to dock chrome or the explicitly enabled picker.
func (s *shell) devPointerHover(x, y int, px, py float64) error {
	if s.dev.resizing || !s.dev.mobile && devOnEdge(s.dev.panel, x, y) {
		s.setCursorShape(ebiten.CursorShapeEWResize)
		return nil
	}
	if devInRect(s.dev.panel, float64(x), float64(y)) {
		s.setCursorShape(ebiten.CursorShapeDefault)
		return nil
	}
	if s.devPickerOn() {
		s.setCursorShape(ebiten.CursorShapeDefault)
		s.devHover(px, py)
		return nil
	}
	return s.hoverCursor(x, y, px, py)
}

// devOpAt returns the topmost painted operation whose box holds the point
// and whose centre sits in the picked box, or -1 when none does. The point
// rule keeps a click on a title, an image, or a row's background on the
// operation the pointer actually touched.
func devOpAt(display *layout.Display, box layout.Box, x, y float64) int {
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
		if x < r.X || y < r.Y || x > r.X+r.W || y > r.Y+r.H {
			continue
		}

		cx, cy := r.X+r.W/devCenterDiv, r.Y+r.H/devCenterDiv
		if cx >= box.X-0.5 && cx <= box.X+box.W+0.5 &&
			cy >= box.Y-0.5 && cy <= box.Y+box.H+0.5 {
			found = i
		}
	}

	return found
}

// devPick handles a press in the content while the overlay is on. Alt
// forwards press, caret, and click to the page; otherwise the press acts on
// the control under the cursor, pinning its box in picker mode.
func (s *shell) devPick(px, py float64, alt bool) error {
	if alt || !s.devPickerOn() {
		s.dev.forward = true

		return host.Batch(s.ctx, s.app, func() error {
			if err := s.ptr().Press(s.ctx, px, py); err != nil {
				return err
			}

			if sel, ok := s.app.(host.Selector); ok {
				if err := sel.SelectAt(s.ctx, px, py); err != nil {
					return err
				}
			}

			return s.ptr().Click(s.ctx, px, py)
		})
	}

	box, ok := devBoxAt(s.app.Boxes(), px, py)
	if !ok {
		s.devDropPin()

		return nil
	}

	if s.dev.havePin && devSameBox(s.dev.pinned, box) {
		s.devDropPin()

		return nil
	}

	s.dev.pinned = box
	s.dev.havePin = true
	s.devShowBoxOps(box, px, py)

	return nil
}

// devPickerOn reports whether the element picker mode is active.
func (s *shell) devPickerOn() bool {
	return s.dev.on && s.dev.picker
}

// devSetPicker changes the explicit picker selection. Turning it off drops
// the hovered and pinned picks and the operation selected through the picker.
func (s *shell) devSetPicker(on bool) {
	s.dev.picker = on && s.dev.on

	if !s.dev.picker {
		s.dev.haveHov = false
		s.dev.havePin = false

		if s.dev.opPin {
			s.dev.haveOp = false
			s.dev.opPin = false
		}
	}
}

// drawDevPick outlines the hovered and pinned boxes, one colour each, with a
// label over each.
func (s *shell) drawDevPick(screen *ebiten.Image) {
	if !s.devPickerOn() {
		return
	}

	if s.dev.haveHov {
		s.devOutline(screen, s.dev.hovered, devHoverInk)
	}

	if s.dev.havePin {
		s.devOutline(screen, s.dev.pinned, devPinInk)
	}
}

func (s *shell) devPickerHit() {
	pw := devTextWidth("Picker")
	_, px := devHeaderRight(s.dev.panel)
	s.dev.hits = append(s.dev.hits, devHit{
		rect: devRect{X: px, Y: s.dev.panel.Y, W: pw, H: devHeaderH},
		act:  devActPicker,
	})
}

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
		cx, cy := r.X+r.W/devCenterDiv, r.Y+r.H/devCenterDiv
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
