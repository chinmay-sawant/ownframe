package ui

import (
	"context"
	"fmt"
	"unicode/utf8"
)

// onPaste applies tab-separated text from the top-left of the selection.
func (a *App) onPaste(_ context.Context, text string) error {
	if a.csv.open {
		return nil
	}

	if a.edit != nil {
		a.edit.insert(text)

		return nil
	}

	rows := decodeTSV(text)
	if len(rows) == 0 {
		return nil
	}

	sh := a.sheet()
	s := a.selection()
	edits := make([]Edit, 0, uiEditCap)
	for i, row := range rows {
		r := s.ActiveR + i
		if r >= sh.Rows {
			break
		}

		for j, field := range row {
			c := s.ActiveC + j
			if c >= sh.Cols || len(edits) >= maxPasteCells {
				break
			}

			edits = append(edits, Edit{Row: r, Col: c, Raw: field})
		}
	}

	if len(edits) == 0 {
		return nil
	}

	a.applyOptimistic(a.active, edits)
	a.postEdits(a.active, edits)
	a.selectRect(s.ActiveR, s.ActiveC, edits[len(edits)-1].Row, edits[len(edits)-1].Col)
	a.status = fmt.Sprintf("pasted %d cells", len(edits))

	return nil
}

// onSelectAll selects every cell, or moves the editor caret to the end.
func (a *App) onSelectAll(_ context.Context) error {
	if a.csv.open {
		return nil
	}

	if a.edit != nil {
		a.edit.Caret = utf8.RuneCountInString(a.edit.Text)

		return nil
	}

	sh := a.sheet()
	a.selectRect(0, 0, sh.Rows-1, sh.Cols-1)

	return nil
}

// onUndo and onRedo queue a history step.
func (a *App) onUndo(_ context.Context) error {
	a.history(true)

	return nil
}

func (a *App) onRedo(_ context.Context) error {
	a.history(false)

	return nil
}
