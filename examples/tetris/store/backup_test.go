package store

import (
	"context"
	"os"
	"path/filepath"
	"testing"
)

func TestBackupAndRestore(t *testing.T) {
	dir := t.TempDir()

	s, err := Open(dir)
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = s.Close() }()
	ctx := context.Background()

	if err := s.SaveGame(ctx, testResult("backed-up", 777), nil); err != nil {
		t.Fatal(err)
	}

	destDir := t.TempDir()
	dest := filepath.Join(destDir, DBName)

	if err := s.Backup(ctx, dest); err != nil {
		t.Fatal(err)
	}

	restored, err := Open(destDir)
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = restored.Close() }()
	top, err := restored.TopScores(ctx, 10)
	if err != nil {
		t.Fatal(err)
	}

	if len(top) != 1 || top[0].ID != "backed-up" || top[0].Score != 777 {
		t.Fatalf("restored scores %v", top)
	}
}

func TestBackupRefusesAnExistingDestination(t *testing.T) {
	dir := t.TempDir()

	s, err := Open(dir)
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = s.Close() }()
	dest := filepath.Join(t.TempDir(), DBName)
	if err := os.WriteFile(dest, []byte("occupied"), 0o600); err != nil {
		t.Fatal(err)
	}

	if err := s.Backup(context.Background(), dest); err == nil {
		t.Fatal("backup overwrote an existing file")
	}
}

func TestBackupRejectsAnEmptyDestination(t *testing.T) {
	s, err := OpenMemory()
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = s.Close() }()
	if err := s.Backup(context.Background(), ""); err == nil {
		t.Fatal("an empty destination was accepted")
	}
}

func TestCheckpointOnMemoryIsANoOp(t *testing.T) {
	s, err := OpenMemory()
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = s.Close() }()
	if err := s.Checkpoint(context.Background()); err != nil {
		t.Fatal(err)
	}
}
