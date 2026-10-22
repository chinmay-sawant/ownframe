package store

import (
	"context"
	"errors"
	"fmt"
	"os"
	"path/filepath"
)

// Backup writes a consistent copy with SQLite VACUUM INTO. Never copy the
// open database file by hand: that misses the WAL and can be torn.
func (s *Store) Backup(ctx context.Context, destPath string) error {
	if destPath == "" {
		return errors.New("store: empty backup path")
	}

	if _, err := os.Lstat(destPath); err == nil {
		return fmt.Errorf("store: backup destination exists: %s", destPath)
	} else if !os.IsNotExist(err) {
		return err
	}

	if err := os.MkdirAll(filepath.Dir(destPath), storeDirPerm); err != nil {
		return err
	}

	if err := s.checkpoint(ctx); err != nil {
		return err
	}

	return s.do(ctx, func(ctx context.Context) error {
		_, err := s.db.ExecContext(ctx, `VACUUM INTO ?`, destPath)

		return err
	})
}

// checkpoint folds the WAL back into the main file when WAL is active. It
// runs on the store worker, never on the UI loop, and a failure only means
// the next reader replays the WAL.
func (s *Store) checkpoint(ctx context.Context) error {
	if s.journal != "wal" {
		return nil
	}

	return s.do(ctx, func(ctx context.Context) error {
		var busy, logFrames, checkpointed int

		err := s.db.QueryRowContext(ctx,
			`PRAGMA wal_checkpoint(PASSIVE)`).Scan(&busy, &logFrames, &checkpointed)
		if err != nil {
			return err
		}

		if busy != 0 {
			// Another reader held the checkpoint; the WAL stays valid.
			return nil
		}

		return nil
	})
}
