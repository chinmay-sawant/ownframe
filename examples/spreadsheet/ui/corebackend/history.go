package corebackend

import (
	"github.com/chinmay-sawant/ownframe/examples/spreadsheet/ui"
)

// Undo reverses the newest step and saves the reversal.
func (b *Backend) Undo() (ui.UndoResult, error) {
	return b.history(true)
}

// Redo reapplies the newest undone step.
func (b *Backend) Redo() (ui.UndoResult, error) {
	return b.history(false)
}

// history moves one step between the two stacks. A save failure puts the
// step back and restores the in-memory workbook.
func (b *Backend) history(undo bool) (ui.UndoResult, error) {
	from, to := &b.undo, &b.redo
	if !undo {
		from, to = &b.redo, &b.undo
	}

	if len(*from) == 0 {
		return ui.UndoResult{Rev: uint64(b.wb.Rev())}, nil //nolint:gosec // reason: revision is a non-negative counter, never exceeds MaxInt64
	}

	s := (*from)[len(*from)-1]
	*from = (*from)[:len(*from)-1]

	cmd, back := s.inv, s.fwd
	if !undo {
		cmd, back = s.fwd, s.inv
	}

	if _, err := b.wb.Apply(cmd); err != nil {
		*from = append(*from, s)

		return ui.UndoResult{Rev: uint64(b.wb.Rev())}, err //nolint:gosec // reason: revision is a non-negative counter, never exceeds MaxInt64
	}

	if err := b.save(cmd); err != nil {
		_, _ = b.wb.Apply(back)
		*from = append(*from, s)

		return ui.UndoResult{Rev: uint64(b.wb.Rev())}, err //nolint:gosec // reason: revision is a non-negative counter, never exceeds MaxInt64
	}

	*to = append(*to, s)

	label := "undo"
	if !undo {
		label = "redo"
	}

	return ui.UndoResult{Rev: uint64(b.wb.Rev()), OK: true, Label: label}, nil //nolint:gosec // reason: revision is a non-negative counter, never exceeds MaxInt64
}
