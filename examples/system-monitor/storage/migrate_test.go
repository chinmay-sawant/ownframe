package storage

import (
	"database/sql"
	"errors"
	"path/filepath"
	"testing"
)

// TestMigrateVersion checks that a fresh file records the schema version and
// that reopening is a no-op.
func TestMigrateVersion(t *testing.T) {
	dir := t.TempDir()

	st, err := Open(dir)
	if err != nil {
		t.Fatal(err)
	}

	version, err := st.SchemaVersionOf(t.Context())
	if err != nil {
		t.Fatal(err)
	}
	if version != SchemaVersion {
		t.Fatalf("version = %d", version)
	}
	if err := st.Close(); err != nil {
		t.Fatal(err)
	}

	st, err = Open(dir)
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = st.Close() }()

	if version, err = st.SchemaVersionOf(t.Context()); err != nil || version != SchemaVersion {
		t.Fatalf("version = %d err = %v", version, err)
	}
}

// TestRejectNewerSchema checks that a file from a newer build is refused and
// left untouched.
func TestRejectNewerSchema(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, dbName)

	raw, err := sql.Open("sqlite", fileDSN(path, false))
	if err != nil {
		t.Fatal(err)
	}
	if _, err := raw.ExecContext(t.Context(), `CREATE TABLE future (x INTEGER); PRAGMA user_version = 99`); err != nil {
		t.Fatal(err)
	}
	_ = raw.Close()

	if _, err := Open(dir); !errors.Is(err, ErrSchemaNewer) {
		t.Fatalf("err = %v, want ErrSchemaNewer", err)
	}

	raw, err = sql.Open("sqlite", fileDSN(path, false))
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = raw.Close() }()

	var version int
	if err := raw.QueryRowContext(t.Context(), "PRAGMA user_version").Scan(&version); err != nil {
		t.Fatal(err)
	}
	if version != 99 {
		t.Fatalf("version changed to %d", version)
	}

	var n int
	if err := raw.QueryRowContext(t.Context(), `SELECT COUNT(*) FROM sqlite_master WHERE name = 'meta'`).Scan(&n); err != nil {
		t.Fatal(err)
	}
	if n != 0 {
		t.Fatal("Open created tables in a newer file")
	}
}

// TestInterruptedMigration is in migrate_fail_test.go.
