package storage

import (
	"context"
	"errors"
	"fmt"
	"os"
	"strings"
)

// Checkpoint folds the WAL back into the database file. The example calls it
// from a worker, never from the UI loop. It is a no-op outside WAL mode.
func (s *Store) Checkpoint(ctx context.Context) error {
	if err := s.ready(); err != nil {
		return err
	}
	if s.journal != "wal" {
		return nil
	}

	ctx, cancel := s.opCtx(ctx)
	defer cancel()

	_, err := s.db.ExecContext(ctx, "PRAGMA wal_checkpoint(PASSIVE)")

	return err
}

// Backup writes a SQLite-aware copy of the database to dest. It uses VACUUM
// INTO, so the copy is consistent even while the database is open, unlike a
// plain file copy of a live WAL database. The destination must not exist.
func (s *Store) Backup(ctx context.Context, dest string) error {
	if err := s.ready(); err != nil {
		return err
	}
	if dest == "" {
		return errors.New("storage: backup destination is empty")
	}
	if _, err := os.Stat(dest); err == nil {
		return fmt.Errorf("storage: backup destination %s exists", dest)
	}

	ctx, cancel := s.opCtx(ctx)
	defer cancel()

	_, err := s.db.ExecContext(ctx, "VACUUM INTO "+quoteSQL(dest)) //nolint:gosec // reason: VACUUM INTO takes no bind parameters; quoteSQL doubles embedded quotes, so the escaped path cannot break out of the literal.

	return err
}

// quoteSQL quotes a string literal for the few statements that cannot take a
// bind parameter, such as VACUUM INTO.
func quoteSQL(s string) string {
	return "'" + strings.ReplaceAll(s, "'", "''") + "'"
}
