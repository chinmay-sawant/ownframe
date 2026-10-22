package store

import (
	"context"
	"database/sql"
)

// open opens one connection pool, migrates, activates the journal, seeds
// the demo scores, and starts the worker. A newer schema is rejected
// before journal activation, so the file stays untouched.
func open(dsn, path string) (*Store, error) {
	db, err := sql.Open("sqlite", dsn)
	if err != nil {
		return nil, err
	}

	db.SetMaxOpenConns(1)
	db.SetMaxIdleConns(1)

	s := &Store{
		db:   db,
		path: path,
		reqs: make(chan *request, QueueSize),
		quit: make(chan struct{}),
		done: make(chan struct{}),
	}

	ctx, cancel := context.WithTimeout(context.Background(), openTimeout)
	defer cancel()

	if err := s.init(ctx); err != nil {
		_ = db.Close()
		return nil, err
	}

	go s.worker()

	return s, nil
}

// init migrates, activates WAL or the rollback fallback, and seeds.
func (s *Store) init(ctx context.Context) error {
	if err := migrate(ctx, s.db); err != nil {
		return err
	}

	mode, err := activateJournal(ctx, s.db)
	if err != nil {
		return err
	}

	s.mode = mode

	return seedIfNeeded(ctx, s.db)
}

// activateJournal tries WAL and falls back to a rollback journal when the
// filesystem cannot host shared memory. The returned name is lowercase.
func activateJournal(ctx context.Context, db *sql.DB) (string, error) {
	var mode string

	err := db.QueryRowContext(ctx, "PRAGMA journal_mode = WAL").Scan(&mode)
	if err == nil && mode == "wal" {
		return mode, nil
	}

	var fallback string
	if ferr := db.QueryRowContext(ctx, "PRAGMA journal_mode = DELETE").Scan(&fallback); ferr != nil {
		if err != nil {
			return "", err
		}

		return "", ferr
	}

	return fallback, nil
}
