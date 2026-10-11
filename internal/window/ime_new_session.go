//go:build android || ios

package window

import (
	"github.com/chinmay-sawant/ownframe/internal/host"
	"github.com/hajimehoshi/ebiten/v2/exp/textinput"
)

// imeNewSession reports the caret and surrounding text for a focused control.
func (s *shell) imeNewSession() *textinput.SessionOptions {
	target, ok := s.app.(host.IME)
	if !ok {
		return nil
	}
	box, caret, before, after, ok := target.IMEContext()
	if !ok {
		return nil
	}
	s.ime.start = caret - len(before)
	s.ime.end = caret + len(after)
	s.ime.caret = s.imeScreenRect(box)
	return &textinput.SessionOptions{CaretBounds: s.ime.caret,
		TextBeforeCaret: before, TextAfterCaret: after}
}
