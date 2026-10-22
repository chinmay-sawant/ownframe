package store

import (
	"context"
	"database/sql"
	"strings"
)

// pruneBatch is the deleted-rows-per-statement bound.
const pruneBatch = 50

// PruneDummy deletes dummy rows outside the best keep in bounded batches,
// at most pruneBatch*10 rows per call. Play never sees a large delete.
func (s *Store) PruneDummy(ctx context.Context, keep int) (int, error) {
	if keep < 0 {
		keep = 0
	}

	total := 0

	err := s.doRetry(ctx, func(ctx context.Context, db *sql.DB) error {
		for round := 0; round < 10; round++ {
			n, err := pruneOnce(ctx, db, keep)
			if err != nil {
				return err
			}

			total += n

			if n < pruneBatch {
				break
			}
		}

		return nil
	})

	return total, err
}

// pruneOnce deletes one batch of dummy rows past the keep rank. It closes
// rows before the delete, because the pool holds one connection.
func pruneOnce(ctx context.Context, db *sql.DB, keep int) (int, error) {
	rows, err := db.QueryContext(ctx, `
		SELECT id FROM scores WHERE dummy = 1
		ORDER BY score DESC, id ASC LIMIT ? OFFSET ?`,
		pruneBatch, keep)
	if err != nil {
		return 0, err
	}

	var ids []string

	for rows.Next() {
		var id string
		if err := rows.Scan(&id); err != nil {
			_ = rows.Close()
			return 0, err
		}

		ids = append(ids, id)
	}

	if err := rows.Err(); err != nil {
		_ = rows.Close()
		return 0, err
	}

	_ = rows.Close()
	if len(ids) == 0 {
		return 0, nil
	}

	args := make([]any, len(ids))
	for i, id := range ids {
		args[i] = id
	}

	res, err := db.ExecContext(ctx,
		`DELETE FROM scores WHERE id IN (?`+strings.Repeat(",?", len(ids)-1)+`)`, //nolint:gosec // reason: placeholders are built from len(ids), no user input enters the SQL string.
		args...)
	if err != nil {
		return 0, err
	}

	n, _ := res.RowsAffected()

	return int(n), nil
}
