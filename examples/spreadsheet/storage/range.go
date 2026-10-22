package storage

import (
	"context"
	"database/sql"

	"github.com/chinmay-sawant/ownframe/examples/spreadsheet/workbook"
)

// LoadRange returns up to limit stored cells inside r, row-major, for one
// sheet. The limit defaults to 5000 and never exceeds 50000, so the UI
// reads bounded tiles instead of whole sheets.
func (s *Store) LoadRange(ctx context.Context, sheet workbook.SheetID, r workbook.Rect, limit int) ([]workbook.CellEdit, error) {
	if limit <= 0 {
		limit = rangeDefLimit
	}

	if limit > rangeMaxLimit {
		limit = rangeMaxLimit
	}

	var out []workbook.CellEdit

	err := s.do(ctx, func(ctx context.Context, db *sql.DB) error {
		rows, err := db.QueryContext(ctx,
			`SELECT row_ix, col_ix, kind, COALESCE(number, 0), COALESCE(text, ''), COALESCE(source, '')
			 FROM cells
			 WHERE sheet_id = ? AND row_ix >= ? AND row_ix <= ? AND col_ix >= ? AND col_ix <= ?
			 ORDER BY row_ix, col_ix LIMIT ?`,
			int64(sheet), r.Min.Row, r.Max.Row, r.Min.Col, r.Max.Col, limit)
		if err != nil {
			return err
		}

		defer func() { _ = rows.Close() }()

		out, err = scanEdits(rows)

		return err
	})

	return out, err
}

// scanEdits reads stored cell rows into edits.
func scanEdits(rows *sql.Rows) ([]workbook.CellEdit, error) {
	var out []workbook.CellEdit

	for rows.Next() {
		var (
			row, col, kind int
			number         float64
			text, source   string
		)

		if err := rows.Scan(&row, &col, &kind, &number, &text, &source); err != nil {
			return nil, err
		}

		c, ok := columnCell(kind, number, text, source)
		if !ok {
			continue
		}

		out = append(out, workbook.CellEdit{
			Pos:  workbook.Pos{Row: row, Col: col},
			Cell: c,
		})
	}

	return out, rows.Err()
}
