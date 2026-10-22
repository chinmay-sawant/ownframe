package store

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strconv"
)

// schemaOf returns the stored schema version. A fresh file is version 0; a
// file with tables but no meta row, or a different app id, is foreign.
func schemaOf(ctx context.Context, db *sql.DB) (int, error) {
	var tables int

	if err := db.QueryRowContext(ctx,
		`SELECT count(*) FROM sqlite_master WHERE type = 'table'`).Scan(&tables); err != nil {
		return 0, fmt.Errorf("%w: %w", ErrCorrupt, err)
	}

	if tables == 0 {
		return 0, nil
	}

	var app string

	err := db.QueryRowContext(ctx, `SELECT value FROM meta WHERE key = 'app_id'`).Scan(&app)
	if errors.Is(err, sql.ErrNoRows) {
		return 0, ErrForeign
	}

	if err != nil {
		return 0, fmt.Errorf("%w: %w", ErrForeign, err)
	}

	if app != appID {
		return 0, ErrForeign
	}

	var raw string

	if err := db.QueryRowContext(ctx,
		`SELECT value FROM meta WHERE key = 'schema_version'`).Scan(&raw); err != nil {
		return 0, fmt.Errorf("%w: missing schema_version", ErrCorrupt)
	}

	v, err := strconv.Atoi(raw)
	if err != nil {
		return 0, fmt.Errorf("%w: bad schema_version %q", ErrCorrupt, raw)
	}

	return v, nil
}
