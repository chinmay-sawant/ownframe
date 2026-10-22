package corebackend

import (
	"context"

	"github.com/chinmay-sawant/ownframe/examples/spreadsheet/ui"
	"github.com/chinmay-sawant/ownframe/examples/spreadsheet/workbook"
)

// undoLimit bounds the adapter's volatile undo history. Undo does not
// survive a restart; the stored cells do.
const undoLimit = 200

// Apply stores an edit batch in memory and on disk. A save failure leaves
// the in-memory workbook as it was, so memory and disk agree again and the
// UI can restore the editor with the typed text.
func (b *Backend) Apply(id string, edits []ui.Edit) (uint64, error) {
	sh, ok := b.sheet(id)
	if !ok {
		return uint64(b.wb.Rev()), ErrNoSheet //nolint:gosec // reason: revision is a non-negative counter, never exceeds MaxInt64
	}

	cmd := workbook.Command{Sheet: sh.ID()}
	for _, e := range edits {
		p := workbook.Pos{Row: e.Row, Col: e.Col}
		if !p.Valid() {
			continue
		}

		cmd.Edits = append(cmd.Edits, workbook.CellEdit{Pos: p, Cell: workbook.ParseInput(e.Raw)})
	}

	inv := b.inverse(sh, cmd.Edits)

	if _, err := b.wb.Apply(cmd); err != nil {
		return uint64(b.wb.Rev()), err //nolint:gosec // reason: revision is a non-negative counter, never exceeds MaxInt64
	}

	if err := b.save(cmd); err != nil {
		_, _ = b.wb.Apply(inv)

		return uint64(b.wb.Rev()), err //nolint:gosec // reason: revision is a non-negative counter, never exceeds MaxInt64
	}

	b.push(step{fwd: cmd, inv: inv})

	return uint64(b.wb.Rev()), nil //nolint:gosec // reason: revision is a non-negative counter, never exceeds MaxInt64
}

// save writes one command through the store and acknowledges only after
// the commit. The workbook's saved revision advances on success.
func (b *Backend) save(cmd workbook.Command) error {
	if len(cmd.Edits) == 0 {
		return nil
	}

	rev, err := b.st.SaveCells(context.Background(), b.wb.ID(), cmd.Sheet, cmd.Edits, b.wb.SavedRev(), "edit")
	if err != nil {
		return err
	}

	b.wb.SetSavedRev(rev)

	return nil
}
