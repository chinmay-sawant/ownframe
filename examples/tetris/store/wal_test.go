package store

import (
	"os"
	"path/filepath"
	"testing"
)

func TestFileStoreActivatesWAL(t *testing.T) {
	dir := t.TempDir()

	s, err := Open(dir)
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = s.Close() }()
	if s.JournalMode() != "wal" {
		t.Fatalf("journal mode %q, want wal", s.JournalMode())
	}

	if _, err := os.Stat(filepath.Join(dir, DBName)); err != nil {
		t.Fatalf("database file: %v", err)
	}
}

func TestOpenCreatesTheDataDirectory(t *testing.T) {
	dir := filepath.Join(t.TempDir(), "nested", "data")

	s, err := Open(dir)
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = s.Close() }()
	if _, err := os.Stat(filepath.Join(dir, DBName)); err != nil {
		t.Fatalf("database file: %v", err)
	}
}
