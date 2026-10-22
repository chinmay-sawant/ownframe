package storage

import (
	"context"
	"database/sql"
	"errors"
	"path/filepath"
	"testing"
)

func TestMigrationFreshAndReopen(t *testing.T) {
	dir := t.TempDir()

	st, err := Open(dir)
	if err != nil {
		t.Fatal(err)
	}

	if err := st.Close(); err != nil {
		t.Fatal(err)
	}

	db, err := sql.Open("sqlite", filepath.Join(dir, File))
	if err != nil {
		t.Fatal(err)
	}

	var version int
	if err := db.QueryRowContext(t.Context(), `PRAGMA user_version`).Scan(&version); err != nil {
		t.Fatal(err)
	}

	if version != SchemaVersion {
		t.Fatalf("user_version = %d, want %d", version, SchemaVersion)
	}

	_ = db.Close()

	st, err = Open(dir)
	if err != nil {
		t.Fatalf("reopen: %v", err)
	}

	if _, err := st.CreateWorkbook(context.Background(), "after reopen"); err != nil {
		t.Fatal(err)
	}

	_ = st.Close()
}

func TestSchemaNewerRejected(t *testing.T) {
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

	if _, err := db.ExecContext(t.Context(), `PRAGMA user_version = 99`); err != nil {
		t.Fatal(err)
	}

	_ = db.Close()

	if _, err := Open(dir); !errors.Is(err, ErrSchemaNewer) {
		t.Fatalf("Open newer schema = %v, want ErrSchemaNewer", err)
	}

	db, err = sql.Open("sqlite", filepath.Join(dir, File))
	if err != nil {
		t.Fatal(err)
	}

	defer func() { _ = db.Close() }()

	var version int
	if err := db.QueryRowContext(t.Context(), `PRAGMA user_version`).Scan(&version); err != nil {
		t.Fatal(err)
	}

	if version != 99 {
		t.Fatalf("version after rejection = %d, want 99 untouched", version)
	}
}
