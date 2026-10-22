package corebackend

import (
	"context"

	"github.com/chinmay-sawant/ownframe/examples/spreadsheet/storage"
	"github.com/chinmay-sawant/ownframe/examples/spreadsheet/ui"
	"github.com/chinmay-sawant/ownframe/examples/spreadsheet/workbook"
)

// CommitCSV applies the whole table in memory and in one storage
// transaction. A failed import reloads the workbook from disk, so the
// stored workbook stays intact and memory matches it again.
func (b *Backend) CommitCSV(id string, data []byte, replace bool) (ui.ImportResult, error) {
	sh, ok := b.sheet(id)
	if !ok {
		return ui.ImportResult{}, ErrNoSheet
	}

	table, err := workbook.ParseCSV(data, workbook.DefaultCSVOptions())
	if err != nil {
		return ui.ImportResult{}, err
	}

	cmd := table.Command(sh.ID(), workbook.Pos{})
	if _, err := b.wb.Apply(cmd); err != nil {
		return ui.ImportResult{}, err
	}

	if _, err := b.st.ImportCells(context.Background(), b.wb.ID(), sh.ID(), cmd.Edits,
		b.wb.SavedRev(), storage.ImportOptions{Replace: replace, Origin: "csv-import"}); err != nil {
		_ = b.reload()

		return ui.ImportResult{}, err
	}

	// Replace clears the stored sheet, so memory is rebuilt from the
	// committed rows instead of patched.
	if err := b.reload(); err != nil {
		return ui.ImportResult{}, err
	}

	return ui.ImportResult{Rev: uint64(b.wb.Rev()), Rows: table.Rows(), Cols: table.Cols()}, nil //nolint:gosec // reason: revision is a non-negative counter, never exceeds MaxInt64
}
