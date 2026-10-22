package store

import (
	"context"
	"database/sql"
	"errors"
	"os"
	"path/filepath"

	_ "modernc.org/sqlite"
)

type Store struct{ db *sql.DB }

// storeDirPerm keeps the database directory owner-only.
const storeDirPerm = 0o700

func DefaultPath() (string, error) {
	dir, err := os.UserConfigDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(dir, "ownframe", "pet-care", "pet-care.db"), nil
}

func Open(path string) (*Store, error) {
	if path != ":memory:" {
		if err := os.MkdirAll(filepath.Dir(path), storeDirPerm); err != nil {
			return nil, err
		}
	}
	db, err := sql.Open("sqlite", path)
	if err != nil {
		return nil, err
	}
	db.SetMaxOpenConns(1)
	if _, err = db.ExecContext(context.Background(), schema); err != nil {
		return nil, errors.Join(err, db.Close())
	}
	return &Store{db: db}, nil
}

func (s *Store) Close() error { return s.db.Close() }

const schema = `PRAGMA foreign_keys=ON;
CREATE TABLE IF NOT EXISTS pets(id TEXT PRIMARY KEY,name TEXT NOT NULL,species TEXT NOT NULL,breed TEXT,birth_date TEXT,notes TEXT);
CREATE TABLE IF NOT EXISTS records(id TEXT PRIMARY KEY,pet_id TEXT NOT NULL REFERENCES pets(id) ON DELETE CASCADE,category TEXT NOT NULL,date TEXT NOT NULL,note TEXT,scheduled INTEGER NOT NULL,complete INTEGER NOT NULL);
CREATE TABLE IF NOT EXISTS app_meta(key TEXT PRIMARY KEY,value TEXT NOT NULL);
PRAGMA user_version=1;`
