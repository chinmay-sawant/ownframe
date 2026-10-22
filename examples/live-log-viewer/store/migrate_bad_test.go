package store

import (
	"bytes"
	"context"
	"database/sql"
	"errors"
	"os"
	"path/filepath"
	"testing"
)

func TestForeignDatabase(t *testing.T) {
	dir := t.TempDir()

	db, err := sql.Open("sqlite", dsn(filepath.Join(dir, dbName)))
	if err != nil {
		t.Fatal(err)
	}

	if _, err := db.ExecContext(context.Background(), `CREATE TABLE other (x TEXT)`); err != nil {
		t.Fatal(err)
	}

	_ = db.Close()

	if _, err := Open(dir); !errors.Is(err, ErrForeign) {
		t.Fatalf("err = %v, want ErrForeign", err)
	}
}

func TestCorruptDatabase(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, dbName)

	if err := os.WriteFile(path, []byte("this is not a sqlite database"), 0o600); err != nil {
		t.Fatal(err)
	}

	before, err := os.ReadFile(path) //nolint:gosec // reason: test reads TempDir database copy to verify it is untouched
	if err != nil {
		t.Fatal(err)
	}

	if _, err := Open(dir); !errors.Is(err, ErrCorrupt) {
		t.Fatalf("err = %v, want ErrCorrupt", err)
	}

	after, err := os.ReadFile(path) //nolint:gosec // reason: test reads TempDir database copy to verify it is untouched
	if err != nil {
		t.Fatal(err)
	}

	if !bytes.Equal(before, after) {
		t.Fatal("a corrupt file was modified")
	}
}

func TestMigrationRollsBack(t *testing.T) {
	db, err := sql.Open("sqlite", dsn(Memory))
	if err != nil {
		t.Fatal(err)
	}

	defer func() { _ = db.Close() }()

	bad := []migration{
		{1, `CREATE TABLE a (x TEXT)`},
		{2, `CREATE TABLE b (`},
	}

	if err := migrateList(bg(), db, bad); err == nil {
		t.Fatal("bad migration returned nil")
	}

	v, err := schemaOf(bg(), db)
	if err != nil || v != 0 {
		t.Fatalf("version = %d, err = %v", v, err)
	}

	var n int

	if err := db.QueryRowContext(t.Context(),
		`SELECT count(*) FROM sqlite_master WHERE name IN ('a', 'b')`).Scan(&n); err != nil {
		t.Fatal(err)
	}

	if n != 0 {
		t.Fatalf("partial schema left %d tables", n)
	}
}
