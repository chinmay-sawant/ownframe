package store

import (
	"bytes"
	"errors"
	"os"
	"path/filepath"
	"testing"
)

func TestRejectNewerSchema(t *testing.T) {
	dir := t.TempDir()

	st, err := Open(dir)
	if err != nil {
		t.Fatal(err)
	}

	if _, err := st.db.ExecContext(bg(),
		`UPDATE meta SET value = '99' WHERE key = 'schema_version'`); err != nil {
		t.Fatal(err)
	}

	if err := st.Checkpoint(bg()); err != nil {
		t.Fatal(err)
	}

	if err := st.Close(); err != nil {
		t.Fatal(err)
	}

	path := filepath.Join(dir, dbName)

	before, err := os.ReadFile(path) //nolint:gosec // reason: test reads TempDir database copy to verify it is untouched
	if err != nil {
		t.Fatal(err)
	}

	if _, err := Open(dir); !errors.Is(err, ErrSchemaNewer) {
		t.Fatalf("err = %v, want ErrSchemaNewer", err)
	}

	after, err := os.ReadFile(path) //nolint:gosec // reason: test reads TempDir database copy to verify it is untouched
	if err != nil {
		t.Fatal(err)
	}

	if !bytes.Equal(before, after) {
		t.Fatal("a newer database was modified")
	}
}
