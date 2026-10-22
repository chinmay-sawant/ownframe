package storage

import (
	"context"
	"database/sql"
	"time"

	"github.com/chinmay-sawant/ownframe/examples/spreadsheet/workbook"
)

// insertWorkbook writes one seeded workbook, its sheets, and its cells.
func insertWorkbook(ctx context.Context, tx *sql.Tx, w *workbook.Workbook) error {
	now := time.Now().UTC().Format(time.RFC3339Nano)

	res, err := tx.ExecContext(ctx,
		`INSERT INTO workbooks (name, rev, created_at, updated_at) VALUES (?, ?, ?, ?)`,
		w.Name(), w.Rev(), now, now)
	if err != nil {
		return err
	}

	id, err := res.LastInsertId()
	if err != nil {
		return err
	}

	w.SetID(workbook.ID(id))

	for pos, sheet := range w.Sheets() {
		res, err := tx.ExecContext(ctx,
			`INSERT INTO sheets (workbook_id, name, position) VALUES (?, ?, ?)`,
			id, sheet.Name(), pos)
		if err != nil {
			return err
		}

		sid, err := res.LastInsertId()
		if err != nil {
			return err
		}

		sheet.SetID(workbook.SheetID(sid))

		if err := insertCells(ctx, tx, sid, sheet); err != nil {
			return err
		}
	}

	return nil
}

// insertCells streams a sheet's cells through one prepared statement.
func insertCells(ctx context.Context, tx *sql.Tx, sheetID int64, sheet *workbook.Sheet) error {
	stmt, err := tx.PrepareContext(ctx,
		`INSERT INTO cells (sheet_id, row_ix, col_ix, kind, number, text, source)
		 VALUES (?, ?, ?, ?, ?, ?, ?)`)
	if err != nil {
		return err
	}

	defer func() { _ = stmt.Close() }()

	var walkErr error

	sheet.Each(func(p workbook.Pos, c workbook.Cell) bool {
		kind, number, text, source := cellColumns(c)

		if _, err := stmt.ExecContext(ctx, sheetID, p.Row, p.Col, kind, number, text, source); err != nil {
			walkErr = err

			return false
		}

		return true
	})

	return walkErr
}
