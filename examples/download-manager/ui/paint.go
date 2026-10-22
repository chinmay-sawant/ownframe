package ui

import (
	"fmt"

	"github.com/chinmay-sawant/ownframe"
)

// paint edits the retained operations for the current view. It opts the
// tick into a dirty-rect replay and marks only the rows that changed, so a
// 1080p frame does not reshape the whole list.
func (a *App) paint() {
	if a.page.Testing().Display() == nil {
		return
	}

	if a.page.Testing().Generation() != a.bound {
		a.bind()
	}

	a.page.Testing().UseFrameDirty()
	rows := a.paintRows()
	a.paintChrome(!rows)
}

// setText writes a retained text run only when the value changed.
func setText(op *ownframe.DisplayOp, text string) bool {
	if op == nil || op.Text == text {
		return false
	}

	op.Text = text

	return true
}

// footText is the measured tick and repaint line.
func (a *App) footText() string {
	s := a.view.Stats

	return fmt.Sprintf("tick %s · paints %d · redraws %d · applied %d",
		s.TickText(), s.Paints, s.Redraws, s.Applied)
}
