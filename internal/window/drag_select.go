package window

import (
	"time"

	"github.com/chinmay-sawant/ownframe/internal/host"
)

// selector is the page range surface the window drives while the pointer is
// down. Deprecated: use host.RangeSelector. A page without it keeps the old
// click behavior.
type selector = host.RangeSelector

// pressAt sends a press, places the caret, then sends the click. The caret
// lands before the click handler runs, so focus and a selection that the
// handler sets are not blurred afterwards. A second press selects the word
// under the point and a third selects the line.
func (s *shell) pressAt(px, py float64) error {
	if claimed, err := s.beginPointerDrag(px, py, true); claimed || err != nil {
		return err
	}
	if err := s.app.Press(s.ctx, px, py); err != nil {
		return err
	}

	sel, ok := s.app.(selector)
	if !ok {
		return s.app.Click(s.ctx, px, py)
	}

	if err := sel.SelectAt(s.ctx, px, py); err != nil {
		return err
	}

	if err := s.app.Click(s.ctx, px, py); err != nil {
		return err
	}

	s.dragActive = true
	s.dragX, s.dragY = px, py

	count := s.clicks.step(time.Now(), px, py)
	if count == 2 {
		return sel.SelectWordAt(s.ctx, px, py)
	}

	if count >= 3 {
		return sel.SelectLineAt(s.ctx, px, py)
	}

	return nil
}

// dragAt extends the range while the button stays down and the point moves.
func (s *shell) dragAt(px, py float64) error {
	if s.gesture.active {
		return s.movePointerDrag(px, py)
	}
	if !s.dragActive || (px == s.dragX && py == s.dragY) {
		return nil
	}

	sel, ok := s.app.(selector)
	if !ok {
		return nil
	}

	s.dragX, s.dragY = px, py

	return sel.Drag(s.ctx, px, py)
}

// releaseAt ends a press and stops the drag.
func (s *shell) releaseAt() error {
	if s.gesture.active {
		return s.endPointerDrag(s.gesture.mouse && !s.gesture.moved)
	}
	s.dragActive = false
	s.hold.release()

	return s.app.Release(s.ctx)
}
