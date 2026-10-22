package store

import (
	"context"
	"os"
	"path/filepath"
	"testing"
)

// TestBackupAndRestore moves a VACUUM INTO copy into a fresh directory.
func TestBackupAndRestore(t *testing.T) {
	s := newStore(t)
	ctx := context.Background()

	if err := s.SaveJob(ctx, testJob("backed-up")); err != nil {
		t.Fatal(err)
	}

	dest := filepath.Join(t.TempDir(), "backup.sqlite")
	if err := s.Backup(ctx, dest); err != nil {
		t.Fatal(err)
	}

	dir := t.TempDir()
	if err := os.Rename(dest, filepath.Join(dir, "jobs.sqlite")); err != nil {
		t.Fatal(err)
	}

	restored, err := Open(dir)
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = restored.Close() }()

	if _, err := restored.Job(ctx, "backed-up"); err != nil {
		t.Errorf("restored row missing: %v", err)
	}
}

// TestBackupRefusesExistingDestination never overwrites a file.
func TestBackupRefusesExistingDestination(t *testing.T) {
	s := newStore(t)

	dest := filepath.Join(t.TempDir(), "backup.sqlite")
	if err := os.WriteFile(dest, []byte("keep"), 0o600); err != nil {
		t.Fatal(err)
	}

	if err := s.Backup(context.Background(), dest); err == nil {
		t.Fatal("backup overwrote an existing file")
	}

	data, _ := os.ReadFile(dest) //nolint:gosec // reason: test reads back the file it wrote.
	if string(data) != "keep" {
		t.Errorf("existing file changed: %q", data)
	}
}
