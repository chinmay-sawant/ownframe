//go:build android || ios

package window

import (
	"unicode/utf8"

	"github.com/hajimehoshi/ebiten/v2/exp/textinput"
)

// imeComposition shows the preedit in the field.
func (s *shell) imeComposition(c *textinput.Composition) {
	s.imeShow(c.Text())
}

// imeShow replaces the shown preedit with text.
func (s *shell) imeShow(text string) {
	if !s.imeTargetCurrent(s.ime.field) || text == s.ime.last {
		return
	}

	s.imeDrop()

	if text != "" {
		if err := s.app.Type(s.ctx, text); err != nil {
			s.ime.err = err
		}
	}

	s.ime.last = text
}

// imeDrop removes the shown preedit from the field.
func (s *shell) imeDrop() {
	if !s.imeTargetCurrent(s.ime.field) {
		s.ime.last = ""
		return
	}
	n := utf8.RuneCountInString(s.ime.last)
	for i := 0; i < n; i++ {
		if err := s.app.Backspace(s.ctx); err != nil {
			s.ime.err = err
		}
	}

	s.ime.last = ""
}
