package storage

import (
	"context"
	"database/sql"

	"github.com/chinmay-sawant/ownframe/examples/spreadsheet/workbook"
)

// writeCells upserts non-blank cells and deletes blank positions. The
// optional progress callback runs every batch and may abort the caller's
// transaction by returning an error.
func writeCells(ctx context.Context, tx *sql.Tx, sheet workbook.SheetID, edits []workbook.CellEdit, progress func(done, total int) error) error {
	const batch = 256

	stmt, err := tx.PrepareContext(ctx,
		`INSERT INTO cells (sheet_id, row_ix, col_ix, kind, number, text, source)
		 VALUES (?, ?, ?, ?, ?, ?, ?)
		 ON CONFLICT (sheet_id, row_ix, col_ix) DO UPDATE SET
		   kind = excluded.kind, number = excluded.number,
		   text = excluded.text, source = excluded.source`)
	if err != nil {
		return err
	}

	defer func() { _ = stmt.Close() }()

	del, err := tx.PrepareContext(ctx,
		`DELETE FROM cells WHERE sheet_id = ? AND row_ix = ? AND col_ix = ?`)
	if err != nil {
		return err
	}

	defer func() { _ = del.Close() }()

	for i, e := range edits {
		if i%batch == 0 {
			if err := ctx.Err(); err != nil {
				return err
			}

			if progress != nil {
				if err := progress(i, len(edits)); err != nil {
					return err
				}
			}
		}

		if e.Cell.Kind == workbook.Blank {
			if _, err := del.ExecContext(ctx, int64(sheet), e.Pos.Row, e.Pos.Col); err != nil {
				return err
			}

			continue
		}

		kind, number, text, source := cellColumns(e.Cell)

		if _, err := stmt.ExecContext(ctx, int64(sheet), e.Pos.Row, e.Pos.Col, kind, number, text, source); err != nil {
			return err
		}
	}

	return nil
}
