package storage

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"github.com/chinmay-sawant/ownframe/examples/spreadsheet/workbook"
)

// LoadWorkbook reads a workbook with all its sheets and cells, then
// recalculates every formula. Calculated caches are never trusted from
// disk, so a reopen cannot serve a stale result.
func (s *Store) LoadWorkbook(ctx context.Context, id workbook.ID) (*workbook.Workbook, error) {
	ctx, cancel := bulkCtx(ctx)
	defer cancel()

	var w *workbook.Workbook

	err := s.do(ctx, func(ctx context.Context, db *sql.DB) error {
		var (
			name string
			rev  int64
		)

		err := db.QueryRowContext(ctx,
			`SELECT name, rev FROM workbooks WHERE id = ?`, int64(id)).Scan(&name, &rev)
		if errors.Is(err, sql.ErrNoRows) {
			return fmt.Errorf("%w: workbook %d", ErrNotFound, id)
		}

		if err != nil {
			return err
		}

		w = workbook.New(id, name)
		w.SetRev(rev)
		w.SetSavedRev(rev)

		sheets, err := loadSheets(ctx, db, id)
		if err != nil {
			return err
		}

		for _, sr := range sheets {
			sheet := w.LoadSheet(sr.id, sr.name)

			if err := loadCells(ctx, db, w, sheet); err != nil {
				return err
			}
		}

		return nil
	})
	if err != nil {
		return nil, err
	}

	w.RecalcAll()

	return w, nil
}

type sheetRow struct {
	id   workbook.SheetID
	name string
}

// loadSheets reads the sheet list and closes the rows before more queries.
func loadSheets(ctx context.Context, db *sql.DB, id workbook.ID) ([]sheetRow, error) {
	rows, err := db.QueryContext(ctx,
		`SELECT id, name FROM sheets WHERE workbook_id = ? ORDER BY position, id`, int64(id))
	if err != nil {
		return nil, err
	}

	defer func() { _ = rows.Close() }()

	var out []sheetRow

	for rows.Next() {
		var (
			sr  sheetRow
			sid int64
		)

		if err := rows.Scan(&sid, &sr.name); err != nil {
			return nil, err
		}

		sr.id = workbook.SheetID(sid)
		out = append(out, sr)
	}

	return out, rows.Err()
}
