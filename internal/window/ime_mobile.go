//go:build android || ios

package window

import (
	"github.com/chinmay-sawant/blinkless/layout"
	"github.com/chinmay-sawant/ownframe/internal/host"
	"github.com/hajimehoshi/ebiten/v2/exp/textinput"
	"image"
	"time"
	"unicode/utf8"
)

// imeState is the platform text input session for the focused control.
type imeState struct {
	composer   textinput.Composer
	field      string
	last       string
	caret      image.Rectangle
	rotation   imeRotationState
	pendingEnd bool
	start      int
	end        int
	handled    bool
	err        error
}

// imeInit wires the composer callbacks.
func (s *shell) imeInit() {
	s.ime.composer.OnNewSession = s.imeNewSession
	s.ime.composer.OnComposition = s.imeComposition
	s.ime.composer.OnCommit = s.imeCommit
	s.ime.composer.OnEndByUser = s.imeEndByUser
}

// imeHandled reports whether the platform input consumed this tick.
func (s *shell) imeHandled() bool { return s.ime.handled }

// imeUpdate runs the text input session for one tick.
func (s *shell) imeUpdate() error {
	id := ""
	if f, ok := s.app.(host.Focuser); ok {
		id = f.FocusID()
	}

	s.ime.err = nil
	if id != s.ime.field {
		if err := s.imeCancelSession(); err != nil {
			return err
		}
		s.ime.pendingEnd = false

		s.ime.field = id
	}
	if s.ime.pendingEnd {
		wait, preserve := s.ime.rotation.resolve(time.Now())
		if wait {
			s.ime.handled = false
			return nil
		}
		s.ime.pendingEnd = false
		if !preserve {
			if f, ok := s.app.(host.Focuser); ok {
				if err := f.Focus(s.ctx, ""); err != nil {
					return err
				}
			}
		}
	}

	s.imeResize()
	handled, err := s.ime.composer.Update()
	s.ime.handled = handled

	if err != nil {
		return err
	}

	return s.ime.err
}

func (s *shell) imeCaretScreenRect(box layout.Box) image.Rectangle {
	return s.imeScreenRect(box)
}

func (s *shell) imeEndByUser() {
	if s.ime.rotation.endByUser(time.Now()) {
		return
	}
	s.ime.last = ""
	s.ime.pendingEnd = true
}

func (s *shell) imeLayoutChanged(w, h int) {
	s.ime.rotation.layout(w, h, time.Now())
}

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
		if err := s.kbd().Type(s.ctx, text); err != nil {
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
		if err := s.kbd().Backspace(s.ctx); err != nil {
			s.ime.err = err
		}
	}

	s.ime.last = ""
}
