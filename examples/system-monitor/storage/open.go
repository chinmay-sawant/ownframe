package storage

import (
	"context"
	"database/sql"

	_ "modernc.org/sqlite"
)

// Open opens or creates the database in dir, running migrations.
func Open(dir string) (*Store, error) {
	return OpenWithOptions(Options{Dir: dir})
}

// OpenWithOptions opens the store described by opts. The pool holds one
// connection, which serializes every statement, so the recorder's worker and
// a history read never interleave inside a transaction.
func OpenWithOptions(opts Options) (*Store, error) {
	if opts.BusyTimeout <= 0 {
		opts.BusyTimeout = defaultBusyTimeout
	}
	if opts.OpTimeout <= 0 {
		opts.OpTimeout = defaultOpTimeout
	}

	path := memoryPath
	if !opts.Temp {
		dir := opts.Dir
		if dir == "" {
			var err error
			if dir, err = DefaultDir(); err != nil {
				return nil, err
			}
		}
		if err := mkdirAll(dir); err != nil {
			return nil, err
		}
		path = joinDir(dir, dbName)
	}

	db, err := sql.Open("sqlite", fileDSN(path, isWindows()))
	if err != nil {
		return nil, err
	}

	db.SetMaxOpenConns(1)
	db.SetMaxIdleConns(1)

	s := &Store{db: db, path: path, opTimeout: opts.OpTimeout, temp: opts.Temp}

	if err := s.applyPragmas(context.Background(), opts); err != nil {
		_ = db.Close()

		return nil, err
	}

	ctx, cancel := context.WithTimeout(context.Background(), migrateTimeout)
	defer cancel()

	if err := s.migrate(ctx); err != nil {
		_ = db.Close()

		return nil, err
	}

	return s, nil
}
