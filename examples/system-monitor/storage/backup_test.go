package storage

import (
	"path/filepath"
	"testing"
	"time"

	"github.com/chinmay-sawant/ownframe/examples/system-monitor/domain"
)

// TestBackupAndRestore checks a SQLite-aware copy: the backup opens as its own
// database with the same rows, and an existing destination is refused.
func TestBackupAndRestore(t *testing.T) {
	st := openTest(t)
	ctx := t.Context()
	id := newTestSession(t, st)

	if _, err := st.Append(ctx, id, []Row{rec(domain.MetricCPU, "", time.Now(), 42)}); err != nil {
		t.Fatal(err)
	}
	if err := st.SetSetting(ctx, "theme", "dark"); err != nil {
		t.Fatal(err)
	}

	dir := t.TempDir()
	dest := filepath.Join(dir, dbName)

	if err := st.Backup(ctx, dest); err != nil {
		t.Fatal(err)
	}
	if err := st.Backup(ctx, dest); err == nil {
		t.Fatal("backup overwrote an existing file")
	}

	restored, err := Open(dir)
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = restored.Close() }()

	if got, _, err := restored.Setting(ctx, "theme"); err != nil || got != "dark" {
		t.Fatalf("theme = %q err = %v", got, err)
	}

	rows, err := restored.History(ctx, Query{SessionID: id, Metric: domain.MetricCPU})
	if err != nil || len(rows) != 1 || rows[0].Value != 42 {
		t.Fatalf("rows = %+v err = %v", rows, err)
	}
}

// TestCheckpoint checks that a checkpoint outside WAL is a no-op and that a
// WAL store checkpoints without error.
func TestCheckpoint(t *testing.T) {
	st := openTest(t)

	if err := st.Checkpoint(t.Context()); err != nil {
		t.Fatal(err)
	}

	rollback, err := OpenWithOptions(Options{Dir: t.TempDir(), ForceRollback: true})
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = rollback.Close() }()

	if err := rollback.Checkpoint(t.Context()); err != nil {
		t.Fatal(err)
	}
}
