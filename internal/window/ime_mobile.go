//go:build android || ios

package window

import (
	"image"
	"time"

	"github.com/hajimehoshi/ebiten/v2/exp/textinput"

	"github.com/chinmay-sawant/ownframe/internal/host"
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
		if s.ime.field != "" {
			s.ime.composer.Cancel()
			s.ime.last = ""
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
