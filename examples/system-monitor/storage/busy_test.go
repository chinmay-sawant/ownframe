package storage

import (
	"context"
	"os"
	"testing"
	"time"
)

// TestLockedDatabase checks that a second store on the same file fails a
// write with a busy error inside its deadline instead of hanging.
func TestLockedDatabase(t *testing.T) {
	dir := t.TempDir()
	ctx := t.Context()

	a, err := OpenWithOptions(Options{Dir: dir, BusyTimeout: 50 * time.Millisecond})
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = a.Close() }()

	b, err := OpenWithOptions(Options{Dir: dir, BusyTimeout: 50 * time.Millisecond})
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = b.Close() }()

	tx, err := a.db.BeginTx(ctx, nil)
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = tx.Rollback() }()

	if _, err := tx.ExecContext(ctx, `INSERT INTO settings(key, value, updated_ns) VALUES('x', '1', 1)`); err != nil {
		t.Fatal(err)
	}

	writeCtx, cancel := context.WithTimeout(ctx, time.Second)
	defer cancel()

	if err := b.SetSetting(writeCtx, "y", "2"); err == nil {
		t.Fatal("write against a locked database succeeded")
	}

	if err := tx.Rollback(); err != nil {
		t.Fatal(err)
	}
	if err := b.SetSetting(ctx, "y", "2"); err != nil {
		t.Fatal(err)
	}
}

// TestReadOnlyDirectory checks the failure mode for unwritable storage.
func TestReadOnlyDirectory(t *testing.T) {
	if os.Geteuid() == 0 {
		t.Skip("root ignores directory permissions")
	}

	dir := t.TempDir()
	if err := os.Chmod(dir, 0o555); err != nil { //nolint:gosec // reason: the test deliberately strips write access to simulate a read-only directory.
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = os.Chmod(dir, 0o755) }) //nolint:gosec // reason: the test restores write access so TempDir cleanup can remove the directory.

	if _, err := Open(dir); err == nil {
		t.Fatal("open succeeded in a read-only directory")
	}
}
