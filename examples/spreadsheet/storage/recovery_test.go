package storage

import (
	"context"
	"testing"

	"github.com/chinmay-sawant/ownframe/examples/spreadsheet/workbook"
)

func TestCrashBetweenEditAndAck(t *testing.T) {
	dir := t.TempDir()
	ctx := context.Background()

	st, err := Open(dir)
	if err != nil {
		t.Fatal(err)
	}

	w, s := newBook(t, st)
	edit(t, st, w, s, workbook.Pos{Row: 0, Col: 0}, "committed")

	// An in-memory edit that never reaches storage.
	cmd := workbook.Command{Sheet: s.ID(), Edits: []workbook.CellEdit{
		{Pos: workbook.Pos{Row: 0, Col: 0}, Cell: workbook.ParseInput("unacknowledged")},
	}}

	if _, err := w.Apply(cmd); err != nil {
		t.Fatal(err)
	}

	_ = st.Close()

	st, err = Open(dir)
	if err != nil {
		t.Fatal(err)
	}

	got, err := st.LoadWorkbook(ctx, w.ID())
	if err != nil {
		t.Fatal(err)
	}

	if v := got.Sheets()[0].Display(workbook.Pos{Row: 0, Col: 0}); v != "committed" {
		t.Fatalf("A1 = %q, an unacknowledged edit survived a crash", v)
	}

	// A commit whose acknowledgement the UI never processed.
	s = got.Sheets()[0]

	cmd = workbook.Command{Sheet: s.ID(), Edits: []workbook.CellEdit{
		{Pos: workbook.Pos{Row: 0, Col: 1}, Cell: workbook.ParseInput("acked in db")},
	}}

	if _, err := got.Apply(cmd); err != nil {
		t.Fatal(err)
	}

	rev, err := st.SaveCells(ctx, got.ID(), s.ID(), cmd.Edits, got.SavedRev(), "lost ack")
	if err != nil {
		t.Fatal(err)
	}

	_ = st.Close()

	st, err = Open(dir)
	if err != nil {
		t.Fatal(err)
	}

	defer func() { _ = st.Close() }()

	got, err = st.LoadWorkbook(ctx, w.ID())
	if err != nil {
		t.Fatal(err)
	}

	if v := got.Sheets()[0].Display(workbook.Pos{Row: 0, Col: 1}); v != "acked in db" {
		t.Fatalf("B1 = %q, committed save was lost", v)
	}

	if got.Rev() != rev {
		t.Fatalf("rev = %d, want %d", got.Rev(), rev)
	}
}
