package store

import (
	"context"
	"database/sql"
	"os"
	"path/filepath"

	"github.com/chinmay-sawant/ownframe/examples/live-log-viewer/entry"
)

// Open opens the database under dir, creating and migrating it when needed.
// An empty dir uses DefaultDir.
func Open(dir string) (*Store, error) { return OpenWithOptions(Options{Dir: dir}) }

// OpenWithOptions opens a database with explicit controls.
func OpenWithOptions(o Options) (*Store, error) {
	dir := o.Dir
	if o.Temp {
		d, err := os.MkdirTemp("", "ownframe-live-log-viewer-")
		if err != nil {
			return nil, err
		}

		dir = d
	} else if dir == "" {
		var err error

		if dir, err = DefaultDir(); err != nil {
			return nil, err
		}
	}

	if dir != Memory {
		if err := os.MkdirAll(dir, storeDirPerm); err != nil {
			return nil, err
		}
	}

	path := Memory
	if dir != Memory {
		path = filepath.Join(dir, dbName)
	}

	ctx, cancel := context.WithTimeout(context.Background(), openTimeout)
	defer cancel()

	db, err := sql.Open("sqlite", dsn(path))
	if err != nil {
		return nil, err
	}

	db.SetMaxOpenConns(1)
	db.SetMaxIdleConns(1)

	if err := db.PingContext(ctx); err != nil {
		_ = db.Close()

		return nil, wrapOpen(err)
	}

	if err := migrate(ctx, db); err != nil {
		_ = db.Close()

		return nil, err
	}

	journal, err := applyJournal(ctx, db)
	if err != nil {
		_ = db.Close()

		return nil, err
	}

	timeout := o.QueryTimeout
	if timeout <= 0 {
		timeout = entry.DefaultPolicy().QueryTimeout
	}

	retention := o.Retention
	if retention == (Retention{}) {
		retention = DefaultRetention()
	}

	s := &Store{
		db: db, dir: dir, path: path, temp: o.Temp, journal: journal,
		timeout: timeout, retention: retention,
		checkpointEvery: o.CheckpointEvery, jobs: make(chan *job, storeJobQueue),
		stop: make(chan struct{}),
	}

	s.wg.Add(1)
	go s.work()

	return s, nil
}
