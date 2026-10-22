package store

import (
	"context"
	"database/sql"
	"errors"
	"os"
	"time"

	_ "modernc.org/sqlite"
)

// openTimeout bounds schema work at startup.
const openTimeout = 10 * time.Second

// open is the shared constructor.
func open(dir string, memory bool) (*Store, error) {
	if !memory {
		if dir == "" {
			return nil, errors.New("store: empty data directory")
		}

		if err := os.MkdirAll(dir, storeDirPerm); err != nil {
			return nil, err
		}
	}

	name, err := dsn(dir, memory)
	if err != nil {
		return nil, err
	}

	db, err := sql.Open("sqlite", name)
	if err != nil {
		return nil, err
	}

	db.SetMaxOpenConns(1)
	db.SetMaxIdleConns(1)

	s := &Store{
		db: db, dir: dir, memory: memory,
		ops:  make(chan func(), storeOpsQueue),
		quit: make(chan struct{}),
		done: make(chan struct{}),
	}
	go s.loop()

	ctx, cancel := context.WithTimeout(context.Background(), openTimeout)
	defer cancel()

	if err := s.do(ctx, s.startup); err != nil {
		_ = s.Close()

		return nil, err
	}

	return s, nil
}
