package store

import (
	"database/sql"
	"errors"
	"path/filepath"
	"strconv"
	"testing"
)

func TestFreshDatabaseMigratesToTheCurrentVersion(t *testing.T) {
	s, err := OpenMemory()
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = s.Close() }()
	got := s.scalar(t, `SELECT value FROM meta WHERE key = 'schema_version'`)
	if got != strconv.Itoa(SchemaVersion) {
		t.Fatalf("schema version %q, want %d", got, SchemaVersion)
	}

	got = s.scalar(t, `SELECT value FROM meta WHERE key = 'seed_version'`)
	if got != strconv.Itoa(SeedVersion) {
		t.Fatalf("seed version %q, want %d", got, SeedVersion)
	}
}

func TestNewerSchemaIsRejectedUntouched(t *testing.T) {
	dir := t.TempDir()

	s, err := Open(dir)
	if err != nil {
		t.Fatal(err)
	}

	if err := s.Close(); err != nil {
		t.Fatal(err)
	}

	path := filepath.Join(dir, DBName)

	raw, err := sql.Open("sqlite", path)
	if err != nil {
		t.Fatal(err)
	}

	if _, err := raw.ExecContext(t.Context(), `UPDATE meta SET value = '99' WHERE key = 'schema_version'`); err != nil {
		t.Fatal(err)
	}

	if err := raw.Close(); err != nil {
		t.Fatal(err)
	}

	if _, err := Open(dir); !errors.Is(err, ErrNewerSchema) {
		t.Fatalf("Open returned %v, want ErrNewerSchema", err)
	}

	raw, err = sql.Open("sqlite", path)
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = raw.Close() }()
	var v string
	if err := raw.QueryRowContext(t.Context(), `SELECT value FROM meta WHERE key = 'schema_version'`).Scan(&v); err != nil {
		t.Fatal(err)
	}

	if v != "99" {
		t.Fatalf("the rejected open rewrote the version to %q", v)
	}
}
