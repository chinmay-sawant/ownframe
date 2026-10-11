//go:build !android && !ios && !js

package window

import (
	"image"

	"github.com/chinmay-sawant/ownframe/internal/host"
	"github.com/hajimehoshi/ebiten/v2/exp/textinput"
)

type imeState struct {
	composer   textinput.Composer
	field      string
	last       string
	caret      image.Rectangle
	start, end int
	handled    bool
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
