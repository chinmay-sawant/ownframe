//go:build !android && !ios && !js

package window

import (
	"github.com/chinmay-sawant/ownframe/internal/host"
	"github.com/hajimehoshi/ebiten/v2/exp/textinput"
)

func (s *shell) imeComposition(c *textinput.Composition) {
	target, ok := s.app.(host.IMEPreeditor)
	if !ok || (c.Text() != "" && !s.imeTargetCurrent(s.ime.field)) {
		return
	}
	start, end := c.SelectionRangeInBytes()
	if err := target.SetIMEComposition(s.ctx, c.Text(), start, end); err != nil {
		s.ime.err = err
	}
}

func (s *shell) imeDrop() {
	target, ok := s.app.(host.IMEPreeditor)
	if !ok {
		return
	}
	if err := target.SetIMEComposition(s.ctx, "", 0, 0); err != nil {
		s.ime.err = err
	}
}
