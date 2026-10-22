package storage

import (
	"context"
	"database/sql"

	"github.com/chinmay-sawant/ownframe/examples/spreadsheet/workbook"
)

// loadCells fills one sheet from the cells table.
func loadCells(ctx context.Context, db *sql.DB, w *workbook.Workbook, sheet *workbook.Sheet) error {
	rows, err := db.QueryContext(ctx,
		`SELECT row_ix, col_ix, kind, COALESCE(number, 0), COALESCE(text, ''), COALESCE(source, '')
		 FROM cells WHERE sheet_id = ? ORDER BY row_ix, col_ix`, int64(sheet.ID()))
	if err != nil {
		return err
	}

	defer func() { _ = rows.Close() }()

	for rows.Next() {
		var (
			row, col, kind int
			number         float64
			text, source   string
		)

		if err := rows.Scan(&row, &col, &kind, &number, &text, &source); err != nil {
			return err
		}

		c, ok := columnCell(kind, number, text, source)
		if !ok {
			continue
		}

		if err := w.SetCell(sheet.ID(), workbook.Pos{Row: row, Col: col}, c); err != nil {
			return err
		}
	}

	return rows.Err()
}
