package storage

import (
	"context"
	"fmt"
)

// applyPragmas sets the pragmas every physical connection needs and records
// the journal mode the file actually uses. The pool holds one connection; if
// it ever grows, these move into the driver DSN.
func (s *Store) applyPragmas(ctx context.Context, opts Options) error {
	ctx, cancel := context.WithTimeout(ctx, pragmaTimeout)
	defer cancel()

	if _, err := s.db.ExecContext(ctx, fmt.Sprintf("PRAGMA busy_timeout = %d", opts.BusyTimeout.Milliseconds())); err != nil {
		return err
	}
	if _, err := s.db.ExecContext(ctx, "PRAGMA foreign_keys = ON"); err != nil {
		return err
	}
	if _, err := s.db.ExecContext(ctx, "PRAGMA synchronous = NORMAL"); err != nil {
		return err
	}

	if opts.Temp {
		s.journal = "memory"

		return nil
	}

	requested := "wal"
	if opts.ForceRollback {
		requested = "delete"
	}

	var got string
	if err := s.db.QueryRowContext(ctx, "PRAGMA journal_mode = "+requested).Scan(&got); err != nil {
		return err
	}
	if got == "" {
		got = "delete"
	}
	s.journal = got

	return nil
}

// opCtx bounds one operation. A caller deadline wins; otherwise the store's
// operation timeout applies.
func (s *Store) opCtx(ctx context.Context) (context.Context, context.CancelFunc) {
	if _, ok := ctx.Deadline(); ok {
		return context.WithCancel(ctx)
	}

	return context.WithTimeout(ctx, s.opTimeout)
}

// ready reports whether the store can be used.
func (s *Store) ready() error {
	if s == nil || s.closed.Load() {
		return ErrClosed
	}

	return nil
}

// boolInt stores a boolean as the 0 or 1 SQLite uses.
func boolInt(v bool) int {
	if v {
		return 1
	}

	return 0
}
