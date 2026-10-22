package storage

import (
	"context"
	"database/sql"

	"github.com/chinmay-sawant/ownframe/examples/spreadsheet/workbook"
)

// WorkbookInfo is one row of the workbook list.
type WorkbookInfo struct {
	ID        workbook.ID
	Name      string
	Rev       int64
	UpdatedAt string
}

// WorkbookPage is one keyset page of workbooks.
type WorkbookPage struct {
	Items []WorkbookInfo
	Next  workbook.ID
	More  bool
}

// ListWorkbooks returns at most size workbooks with an ID after cursor,
// ordered by ID, initially 50 per page. Cursor zero starts at the first.
func (s *Store) ListWorkbooks(ctx context.Context, cursor workbook.ID, size int) (WorkbookPage, error) {
	if size <= 0 || size > workbook.PageSize {
		size = workbook.PageSize
	}

	var page WorkbookPage

	err := s.do(ctx, func(ctx context.Context, db *sql.DB) error {
		rows, err := db.QueryContext(ctx,
			`SELECT id, name, rev, updated_at FROM workbooks WHERE id > ? ORDER BY id LIMIT ?`,
			int64(cursor), size+1)
		if err != nil {
			return err
		}

		defer func() { _ = rows.Close() }()

		for rows.Next() {
			var info WorkbookInfo
			var id int64

			if err := rows.Scan(&id, &info.Name, &info.Rev, &info.UpdatedAt); err != nil {
				return err
			}

			info.ID = workbook.ID(id)
			page.Items = append(page.Items, info)
		}

		if err := rows.Err(); err != nil {
			return err
		}

		if len(page.Items) > size {
			page.Items = page.Items[:size]
			page.More = true
		}

		if n := len(page.Items); n > 0 {
			page.Next = page.Items[n-1].ID
		}

		return nil
	})

	return page, err
}
