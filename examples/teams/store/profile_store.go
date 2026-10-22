package store

import (
	"context"
	"database/sql"
)

// saveProfile replaces the profile row with the given state.
func saveProfile(tx *sql.Tx, presence string, dark bool) error {
	_, err := tx.ExecContext(context.Background(), `INSERT OR REPLACE INTO profile (id, presence, dark) VALUES (1, ?, ?)`,
		presence, dark)

	return err
}

// loadProfile reads the profile row. A missing row is online and dark.
func loadProfile(db *sql.DB) (presence string, dark bool, err error) {
	presence, dark = "online", true

	err = db.QueryRowContext(context.Background(), `SELECT presence, dark FROM profile WHERE id = 1`).Scan(&presence, &dark)
	if err != nil && err != sql.ErrNoRows {
		return "", false, err
	}

	return presence, dark, nil
}

// readInto runs query and appends one scanned value per row.
func readInto[T any](db *sql.DB, query string, out *[]T, scan func(*sql.Rows, *T) error) error {
	rows, err := db.QueryContext(context.Background(), query)
	if err != nil {
		return err
	}
	defer func() { _ = rows.Close() }()

	for rows.Next() {
		var v T

		if err := scan(rows, &v); err != nil {
			return err
		}

		*out = append(*out, v)
	}

	return rows.Err()
}
