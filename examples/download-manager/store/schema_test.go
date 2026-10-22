package store

import (
	"context"
	"database/sql"
	"errors"
	"os"
	"path/filepath"
	"testing"
)

// TestNewerSchemaRejected leaves the file untouched.
func TestNewerSchemaRejected(t *testing.T) {
	dir := t.TempDir()
	ctx := context.Background()

	s, err := Open(dir)
	if err != nil {
		t.Fatal(err)
	}

	if _, err := s.db.ExecContext(ctx,
		`INSERT INTO schema_migrations (version, applied_ms) VALUES (99, 0)`); err != nil {
		t.Fatal(err)
	}

	if err := s.SaveJob(ctx, testJob("before")); err != nil {
		t.Fatal(err)
	}

	if err := s.Close(); err != nil {
		t.Fatal(err)
	}

	_, err = Open(dir)
	if !errors.Is(err, ErrNewerSchema) {
		t.Fatalf("want ErrNewerSchema, got %v", err)
	}

	name, err := dsn(dir, false)
	if err != nil {
		t.Fatal(err)
	}

	db, err := sql.Open("sqlite", name)
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = db.Close() }()

	var max int
	if err := db.QueryRow(`SELECT max(version) FROM schema_migrations`).Scan(&max); err != nil {
		t.Fatal(err)
	}

	if max != 99 {
		t.Errorf("schema version changed to %d", max)
	}

	var jobs int
	if err := db.QueryRow(`SELECT count(*) FROM jobs`).Scan(&jobs); err != nil {
		t.Fatal(err)
	}

	if jobs != 1 {
		t.Errorf("jobs count changed to %d", jobs)
	}
}

// TestCorruptDatabaseFailsWithoutReset reports a clear open error.
func TestCorruptDatabaseFailsWithoutReset(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "jobs.sqlite")

	if err := os.WriteFile(path, []byte("not a database"), 0o600); err != nil {
		t.Fatal(err)
	}

	if _, err := Open(dir); err == nil {
		t.Fatal("corrupt database opened")
	}

	data, _ := os.ReadFile(path) //nolint:gosec // reason: test reads back the corrupt file it wrote.
	if string(data) != "not a database" {
		t.Error("corrupt file was rewritten")
	}
}
