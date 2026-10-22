package store

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"os"
	"strings"
)

// Backup writes a SQLite-aware copy of the database to dest with VACUUM
// INTO. It refuses to overwrite an existing file, and the copy is a single
// consistent file that needs no sidecar journal.
func (s *Store) Backup(ctx context.Context, dest string) error {
	if s.path == Memory {
		return errors.New("store: cannot back up an in-memory database")
	}

	_, err := runJobLong(s, ctx, func(ctx context.Context, db *sql.DB) (int, error) {
		if _, err := os.Stat(dest); err == nil {
			return 0, fmt.Errorf("store: backup destination exists: %s", dest)
		}

		quoted := strings.ReplaceAll(dest, "'", "''")
		if _, err := db.ExecContext(ctx, `VACUUM INTO '`+quoted+`'`); err != nil { //nolint:gosec // reason: VACUUM INTO quotes the destination by doubling single quotes, no placeholders exist
			return 0, err
		}

		return 1, nil
	})

	return err
}
