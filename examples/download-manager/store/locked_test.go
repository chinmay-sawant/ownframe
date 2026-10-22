package store

import (
	"context"
	"database/sql"
	"testing"
	"time"
)

// TestLockedDatabaseRespectsDeadline waits for the busy timeout but gives
// up at the caller's deadline, and never replays the write blindly.
func TestLockedDatabaseRespectsDeadline(t *testing.T) {
	dir := t.TempDir()

	s, err := Open(dir)
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = s.Close() }()

	name, err := dsn(dir, false)
	if err != nil {
		t.Fatal(err)
	}

	holder, err := sql.Open("sqlite", name)
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = holder.Close() }()

	holder.SetMaxOpenConns(1)

	tx, err := holder.BeginTx(context.Background(), nil)
	if err != nil {
		t.Fatal(err)
	}

	if _, err := tx.ExecContext(context.Background(), `UPDATE meta SET value = 'held' WHERE key = 'missing'`); err != nil {
		t.Fatal(err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 200*time.Millisecond)
	defer cancel()

	start := time.Now()

	if err := s.SaveJob(ctx, testJob("blocked")); err == nil {
		t.Fatal("write succeeded under a foreign lock")
	}

	if elapsed := time.Since(start); elapsed > 2*time.Second {
		t.Errorf("deadline ignored: waited %v", elapsed)
	}

	if err := tx.Rollback(); err != nil {
		t.Fatal(err)
	}
}
