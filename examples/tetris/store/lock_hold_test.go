package store

import (
	"context"
	"database/sql"
	"path/filepath"
	"testing"
	"time"
)

func TestSaveGameGivesUpWhenTheLockHolds(t *testing.T) {
	dir := t.TempDir()

	s, err := Open(dir)
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = s.Close() }()
	raw, err := sql.Open("sqlite",
		"file:"+slashAbs(filepath.Join(dir, DBName))+"?_pragma=busy_timeout(100)")
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = raw.Close() }()
	raw.SetMaxOpenConns(1)

	tx, err := raw.BeginTx(context.Background(), nil)
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = tx.Rollback() }()
	if _, err := tx.ExecContext(context.Background(), `UPDATE scores SET score = score WHERE id = 'dummy-0001'`); err != nil {
		t.Fatal(err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 300*time.Millisecond)
	defer cancel()

	start := time.Now()

	if err := s.SaveGame(ctx, testResult("never", 1), nil); err == nil {
		t.Fatal("a save under a held lock succeeded")
	}

	if time.Since(start) > 8*time.Second {
		t.Fatal("the save did not respect its deadline")
	}
}
