//go:build android || ios

package window

import (
	"time"

	"github.com/hajimehoshi/ebiten/v2/exp/textinput"

	"github.com/chinmay-sawant/ownframe/internal/host"
)

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

	if err := s.app.Type(s.ctx, c.Text()); err != nil {
		s.ime.err = err
	}
}

// imeEndByUser waits for a rotation resize before clearing field focus.
func (s *shell) imeEndByUser() {
	if s.ime.rotation.endByUser(time.Now()) {
		return
	}
	s.ime.last = ""
	s.ime.pendingEnd = true
}
