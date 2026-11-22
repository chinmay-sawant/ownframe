//go:build !android && !ios && !js

package window

import (
	"github.com/chinmay-sawant/blinkless/layout"
	"github.com/chinmay-sawant/ownframe/internal/host"
	"github.com/hajimehoshi/ebiten/v2/exp/textinput"
	"image"
)

type imeState struct {
	composer   textinput.Composer
	field      string
	last       string
	caret      image.Rectangle
	start, end int
	handled    bool
	composing  bool
	err        error
}

func (s *shell) imeInit() {
	s.ime.composer.OnNewSession = s.imeNewSession
	s.ime.composer.OnComposition = s.imeComposition
	s.ime.composer.OnCommit = s.imeCommit
	s.ime.composer.OnEndByUser = s.imeEndByUser
}

func (s *shell) imeHandled() bool { return s.ime.handled }

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
		s.ime.field = id
	}
	s.imeResize()
	handled, err := s.ime.composer.Update()
	s.ime.handled = handled
	if err != nil {
		return err
	}
	return s.ime.err
}

func (s *shell) imeLayoutChanged(int, int) {}

func (s *shell) imeCaretScreenRect(box layout.Box) image.Rectangle {
	if target, ok := s.app.(host.IMECaretAnchor); ok {
		if anchor, ok := target.IMEAnchor(); ok {
			box = anchor
		}
	}
	return s.imeScreenRect(box)
}

// imeEndByUser leaves field focus intact after a desktop input method closes.
func (s *shell) imeEndByUser() {}

func (s *shell) imeComposition(c *textinput.Composition) {
	start, end := c.SelectionRangeInBytes()
	s.imePreedit(c.Text(), start, end)
}

// imePreedit forwards the platform's composition to the screen. It tracks
// whether a composition is live so imeResize does not confirm it midway.
func (s *shell) imePreedit(text string, start, end int) {
	target, ok := s.app.(host.IMEPreeditor)
	if !ok || (text != "" && !s.imeTargetCurrent(s.ime.field)) {
		return
	}
	s.ime.composing = text != ""
	if err := target.SetIMEComposition(s.ctx, text, start, end); err != nil {
		s.ime.err = err
	}
}

func (s *shell) imeDrop() {
	s.ime.composing = false
	target, ok := s.app.(host.IMEPreeditor)
	if !ok {
		return
	}
	if err := target.SetIMEComposition(s.ctx, "", 0, 0); err != nil {
		s.ime.err = err
	}
}
