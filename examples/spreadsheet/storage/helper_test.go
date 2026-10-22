package storage

import (
	"context"
	"database/sql"
	"path/filepath"
	"testing"

	"github.com/chinmay-sawant/ownframe/examples/spreadsheet/workbook"
)

func openMemory(t *testing.T) *Store {
	t.Helper()

	st, err := Open(Memory)
	if err != nil {
		t.Fatal(err)
	}

	t.Cleanup(func() { _ = st.Close() })

	return st
}

func openDir(t *testing.T) (*Store, string) {
	t.Helper()

	dir := t.TempDir()

	st, err := Open(dir)
	if err != nil {
		t.Fatal(err)
	}

	t.Cleanup(func() { _ = st.Close() })

	return st, dir
}

// raw opens a second connection to the same file for inspection.
func raw(t *testing.T, dir string) *sql.DB {
	t.Helper()

	db, err := sql.Open("sqlite", filepath.Join(dir, File))
	if err != nil {
		t.Fatal(err)
	}

	t.Cleanup(func() { _ = db.Close() })

	return db
}

// newBook creates a workbook and returns it with its first sheet.
func newBook(t *testing.T, st *Store) (*workbook.Workbook, *workbook.Sheet) {
	t.Helper()

	w, err := st.CreateWorkbook(context.Background(), "Book")
	if err != nil {
		t.Fatal(err)
	}

	return w, w.Sheets()[0]
}

// edit applies a command in memory and saves it, returning the new rev.
func edit(t *testing.T, st *Store, w *workbook.Workbook, s *workbook.Sheet, p workbook.Pos, v string) int64 {
	t.Helper()

	cmd := workbook.Command{Sheet: s.ID(), Edits: []workbook.CellEdit{
		{Pos: p, Cell: workbook.ParseInput(v)},
	}}

	if _, err := w.Apply(cmd); err != nil {
		t.Fatal(err)
	}

	rev, err := st.SaveCells(context.Background(), w.ID(), s.ID(), cmd.Edits, w.SavedRev(), "test")
	if err != nil {
		t.Fatal(err)
	}

	w.SetSavedRev(rev)

	return rev
}
