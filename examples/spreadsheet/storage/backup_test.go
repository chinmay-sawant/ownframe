package storage

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	"github.com/chinmay-sawant/ownframe/examples/spreadsheet/workbook"
)

func TestBackupAndRestore(t *testing.T) {
	st, dir := openDir(t)
	ctx := context.Background()

	w, s := newBook(t, st)
	edit(t, st, w, s, workbook.Pos{Row: 0, Col: 0}, "backed up")

	dest := filepath.Join(t.TempDir(), "backup.db")

	if err := st.Backup(ctx, dest); err != nil {
		t.Fatal(err)
	}

	if err := st.Backup(ctx, dest); err == nil {
		t.Fatal("backup over an existing destination succeeded")
	}

	_ = st.Close()

	restoreDir := t.TempDir()

	if err := os.Rename(dest, filepath.Join(restoreDir, File)); err != nil {
		t.Fatal(err)
	}

	st, err := Open(restoreDir)
	if err != nil {
		t.Fatal(err)
	}

	defer func() { _ = st.Close() }()

	got, err := st.LoadWorkbook(ctx, w.ID())
	if err != nil {
		t.Fatal(err)
	}

	if v := got.Sheets()[0].Display(workbook.Pos{Row: 0, Col: 0}); v != "backed up" {
		t.Fatalf("restored A1 = %q", v)
	}

	if got.Rev() != 1 {
		t.Fatalf("restored rev = %d, want 1", got.Rev())
	}

	if _, err := os.Stat(dir); err != nil {
		t.Fatal(err)
	}
}
