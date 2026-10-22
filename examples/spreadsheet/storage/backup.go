package storage

import (
	"context"
	"database/sql"
	"errors"
	"strings"
)

// Backup writes a consistent copy of the database with VACUUM INTO.
// Copying the file alone can miss committed WAL pages. dest must not
// exist. On Windows, close the source database before renaming or deleting
// the destination, and never delete an active -wal or -shm file.
func (s *Store) Backup(ctx context.Context, dest string) error {
	if dest == "" {
		return errors.New("storage: empty backup destination")
	}

	ctx, cancel := bulkCtx(ctx)
	defer cancel()

	return s.do(ctx, func(ctx context.Context, db *sql.DB) error {
		_, err := db.ExecContext(ctx, `VACUUM INTO '`+strings.ReplaceAll(dest, "'", "''")+`'`) //nolint:gosec // reason: VACUUM INTO quotes the destination by doubling single quotes, no placeholders exist

		return err
	})
}
