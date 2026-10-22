package storage

import (
	"os"
	"path/filepath"
	"testing"
)

// TestCorruptFile checks that a file that is not a database fails Open and is
// left exactly as it was, never reset.
func TestCorruptFile(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, dbName)

	garbage := []byte("this is not a sqlite database, it is plain text\n")

	if err := os.WriteFile(path, garbage, 0o600); err != nil {
		t.Fatal(err)
	}

	if _, err := Open(dir); err == nil {
		t.Fatal("corrupt file opened")
	}

	got, err := os.ReadFile(path) //nolint:gosec // reason: the test reads back the corrupt file it just wrote under TempDir.
	if err != nil {
		t.Fatal(err)
	}
	if string(got) != string(garbage) {
		t.Fatal("Open modified a corrupt file")
	}
}

// TestEmptyFile checks that a zero length file, which SQLite treats as an
// empty database, migrates normally.
func TestEmptyFile(t *testing.T) {
	dir := t.TempDir()

	if err := os.WriteFile(filepath.Join(dir, dbName), nil, 0o600); err != nil {
		t.Fatal(err)
	}

	st, err := Open(dir)
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = st.Close() }()

	if version, err := st.SchemaVersionOf(t.Context()); err != nil || version != SchemaVersion {
		t.Fatalf("version = %d err = %v", version, err)
	}
}
