package window

import (
	"github.com/chinmay-sawant/ownframe/internal/host"
)

// devPick handles a press in the content while the overlay is on. Alt
// forwards press, caret, and click to the page; otherwise the press acts on
// the control under the cursor, pinning its box in picker mode.
func (s *shell) devPick(px, py float64, alt bool) error {
	if alt || !s.devPickerOn() {
		s.dev.forward = true

		if err := s.ptr().Press(s.ctx, px, py); err != nil {
			return err
		}

		if sel, ok := s.app.(host.Selector); ok {
			if err := sel.SelectAt(s.ctx, px, py); err != nil {
				return err
			}
		}

		return s.ptr().Click(s.ctx, px, py)
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
