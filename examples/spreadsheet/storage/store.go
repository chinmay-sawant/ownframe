// Package storage persists spreadsheet workbooks in SQLite through
// database/sql and modernc.org/sqlite. One worker goroutine owns the single
// connection, so calls serialize, migrations are transactional, and saves
// are acknowledged only after COMMIT. The default location is
// os.UserConfigDir()/ownframe/spreadsheet/spreadsheet.db and Open takes an
// explicit directory. Lock contention waits inside SQLite's busy_timeout;
// this package never replays a non-idempotent write on its own. Nothing
// here imports the UI.
package storage

import (
	"context"
	"database/sql"
	"os"
	"path/filepath"
	"sync"
	"sync/atomic"
	"time"

	_ "modernc.org/sqlite"
)

// Store is one open database with one serialized worker and one connection.
type Store struct {
	db      *sql.DB
	path    string
	reqs    chan request
	quit    chan struct{}
	done    chan struct{}
	closed  atomic.Bool
	timeout time.Duration

	journalMu sync.Mutex
	journal   string
}

type request struct {
	ctx  context.Context
	fn   func(context.Context, *sql.DB) error
	done chan error
}

// Open opens the database in dir, creating the directory and the schema
// when needed. An empty dir means DefaultDir; Memory stays in RAM.
func Open(dir string) (*Store, error) {
	if dir == "" {
		d, err := DefaultDir()
		if err != nil {
			return nil, err
		}

		dir = d
	}

	memory := dir == Memory
	path := dir

	if !memory {
		if err := os.MkdirAll(dir, storeDirPerm); err != nil {
			return nil, err
		}

		path = filepath.Join(dir, File)
	}

	db, err := sql.Open("sqlite", dsn(path, memory))
	if err != nil {
		return nil, err
	}

	db.SetMaxOpenConns(1)
	db.SetMaxIdleConns(1)

	s := &Store{
		db:      db,
		path:    path,
		reqs:    make(chan request, storeQueueCap),
		quit:    make(chan struct{}),
		done:    make(chan struct{}),
		timeout: storeTimeout,
	}

	go s.worker()

	if err := s.migrate(context.Background()); err != nil {
		_ = s.Close()

		return nil, err
	}

	s.setJournal(s.readJournal())

	return s, nil
}
