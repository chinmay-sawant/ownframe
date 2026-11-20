//go:build !js

package window

import (
	"github.com/chinmay-sawant/ownframe/internal/host"
	"github.com/hajimehoshi/ebiten/v2/exp/textinput"
	"image"
)

func (s *shell) imeCancelSession() error {
	s.ime.err = nil
	s.ime.composer.Cancel()
	s.ime.field, s.ime.last = "", ""
	return s.ime.err
}

func (s *shell) imeShutdown() { _ = s.imeCancelSession() }

// imeCommit applies a committed edit.
func (s *shell) imeCommit(c *textinput.Commit) {
	if !s.imeTargetCurrent(s.ime.field) {
		return
	}
	s.imeDrop()

	if before, after := c.IsSurroundingTextReplaced(); before || after {
		if target, ok := s.app.(host.IME); ok {
			prefix, suffix := c.SurroundingText()
			text := c.Text()
			caret := len(prefix) + len(text)
			err := target.IMEReplace(s.ctx, s.ime.start, s.ime.end, prefix+text+suffix, caret)
			if err != nil {
				s.ime.err = err
			}
		}

		return
	}

	if c.Text() == "" {
		return
	}

	if err := s.kbd().Type(s.ctx, c.Text()); err != nil {
		s.ime.err = err
	}
}

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
	s.ime.caret = s.imeCaretScreenRect(box)
	return &textinput.SessionOptions{CaretBounds: s.ime.caret,
		TextBeforeCaret: before, TextAfterCaret: after}
}

// imeResize restarts the session when the field box moves because Ebiten
// keeps the caret bounds from the session start.
func (s *shell) imeResize() {
	target, ok := s.app.(host.IME)
	if !ok || s.ime.field == "" {
		return
	}

	box, _, _, _, ok := target.IMEContext()
	if !ok || imeNear(s.imeCaretScreenRect(box), s.ime.caret) {
		return
	}

	s.ime.composer.Confirm()
	s.ime.last = ""
}

// imeNear reports whether two caret boxes are close enough that the
// rasterizer's rounding should not restart the session. Only a real layout
// move, such as the keyboard opening, is worth a restart.
func imeNear(a, b image.Rectangle) bool {
	const slack = 4

	return abs(a.Min.X-b.Min.X) <= slack && abs(a.Min.Y-b.Min.Y) <= slack &&
		abs(a.Max.X-b.Max.X) <= slack && abs(a.Max.Y-b.Max.Y) <= slack
}

func abs(n int) int {
	if n < 0 {
		return -n
	}

	return n
}
