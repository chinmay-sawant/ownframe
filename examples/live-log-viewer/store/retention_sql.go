package store

import (
	"context"
	"database/sql"
)

func entryStats(ctx context.Context, db *sql.DB) (int64, int64, error) {
	var count, bytes int64

	err := db.QueryRowContext(ctx,
		`SELECT count(*), COALESCE(SUM(bytes), 0) FROM entries`).
		Scan(&count, &bytes)

	return count, bytes, err
}

func deleteWhere(ctx context.Context, db *sql.DB, where string, arg any) (int64, error) {
	//nolint:gosec // reason: where clause is built from internal placeholders in whereSQL, values are bound args
	res, err := db.ExecContext(ctx,
		`DELETE FROM entries WHERE id IN
		 (SELECT id FROM entries WHERE `+where+` ORDER BY id LIMIT ?)`,
		arg, pruneBatch)
	if err != nil {
		return 0, err
	}

	return res.RowsAffected()
}

func deleteOldest(ctx context.Context, db *sql.DB, batch int64) (int64, error) {
	res, err := db.ExecContext(ctx,
		`DELETE FROM entries WHERE id IN
		 (SELECT id FROM entries ORDER BY id LIMIT ?)`, batch)
	if err != nil {
		return 0, err
	}

	return res.RowsAffected()
}
