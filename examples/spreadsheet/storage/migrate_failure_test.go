package storage

import (
	"bytes"
	"context"
	"database/sql"
	"os"
	"path/filepath"
	"testing"
)

func TestInterruptedMigration(t *testing.T) {
	dir := t.TempDir()

	st, err := Open(dir)
	if err != nil {
		t.Fatal(err)
	}

	_ = st.Close()

	db, err := sql.Open("sqlite", filepath.Join(dir, File))
	if err != nil {
		t.Fatal(err)
	}

	defer func() { _ = db.Close() }()

	migs := []migration{
		{version: 1, name: "one", sql: `CREATE TABLE t1 (x INTEGER)`},
		{version: 2, name: "bad", sql: `CREATE TABLE t2 (x INTEGER`},
	}

	if err := applyMigrations(context.Background(), db, 2, migs); err == nil {
		t.Fatal("bad migration succeeded")
	}

	var version int
	if err := db.QueryRowContext(t.Context(), `PRAGMA user_version`).Scan(&version); err != nil {
		t.Fatal(err)
	}

	if version != 1 {
		t.Fatalf("version = %d, want 1 after rollback", version)
	}

	var n int
	if err := db.QueryRowContext(t.Context(), `SELECT COUNT(*) FROM sqlite_master WHERE name = 't2'`).Scan(&n); err != nil {
		t.Fatal(err)
	}

	if n != 0 {
		t.Fatal("failed migration left a table behind")
	}
}

func TestCorruptDatabaseUntouched(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, File)

	if err := os.WriteFile(path, []byte("this is not a sqlite database"), 0o600); err != nil {
		t.Fatal(err)
	}

	before, err := os.ReadFile(path) //nolint:gosec // reason: test reads TempDir database copy to verify it is untouched
	if err != nil {
		t.Fatal(err)
	}

	if _, err := Open(dir); err == nil {
		t.Fatal("Open on a corrupt file succeeded")
	}

	after, err := os.ReadFile(path) //nolint:gosec // reason: test reads TempDir database copy to verify it is untouched
	if err != nil {
		t.Fatal(err)
	}

	if !bytes.Equal(before, after) {
		t.Fatal("Open modified the corrupt database file")
	}
}
