// Package store keeps the Teams example state in SQLite. Open creates the
// tables on first use, Seed fills them from the sample data, and Save
// writes the whole state back, so a change survives a restart. A missing or
// unreadable database is not fatal: the app falls back to the sample data.
package store

import (
	"context"
	"database/sql"
	"os"
	"path/filepath"

	_ "modernc.org/sqlite"
)

// Memory is a database that lasts as long as the process. Tests use it.
const Memory = ":memory:"

// Store is one open SQLite database.
type Store struct {
	db *sql.DB
}

// Open opens path, creating the parent directory and the schema when
// needed.
func Open(path string) (*Store, error) {
	if path != Memory {
		if err := os.MkdirAll(filepath.Dir(path), storeDirPerm); err != nil {
			return nil, err
		}
	}

	db, err := sql.Open("sqlite", path)
	if err != nil {
		return nil, err
	}

	db.SetMaxOpenConns(1)

	if _, err := db.ExecContext(context.Background(), schema); err != nil {
		_ = db.Close()

		return nil, err
	}

	return &Store{db: db}, nil
}

// Close closes the database.
func (s *Store) Close() error {
	if s == nil || s.db == nil {
		return nil
	}

	return s.db.Close()
}
