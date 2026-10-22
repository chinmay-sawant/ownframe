package storage

import (
	"database/sql"
	"path/filepath"
	"testing"
)

// TestInterruptedMigration checks that a failed migration rolls back: no
// partial schema and no version bump.
func TestInterruptedMigration(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, dbName)

	raw, err := sql.Open("sqlite", fileDSN(path, false))
	if err != nil {
		t.Fatal(err)
	}
	// A table with this name makes the schema's index creation fail.
	if _, err := raw.Exec(`CREATE TABLE metric_history (other INTEGER)`); err != nil {
		t.Fatal(err)
	}
	_ = raw.Close()

	if _, err := Open(dir); err == nil {
		t.Fatal("migration succeeded against a conflicting table")
	}

	raw, err = sql.Open("sqlite", fileDSN(path, false))
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = raw.Close() }()

	var version int
	if err := raw.QueryRow("PRAGMA user_version").Scan(&version); err != nil {
		t.Fatal(err)
	}
	if version != 0 {
		t.Fatalf("version = %d after a failed migration", version)
	}

	var n int
	if err := raw.QueryRow(`SELECT COUNT(*) FROM sqlite_master WHERE name = 'meta'`).Scan(&n); err != nil {
		t.Fatal(err)
	}
	if n != 0 {
		t.Fatal("failed migration left a partial schema")
	}
}
