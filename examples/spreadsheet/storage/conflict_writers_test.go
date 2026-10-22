package storage

import (
	"context"
	"strconv"
	"testing"

	"github.com/chinmay-sawant/ownframe/examples/spreadsheet/workbook"
)

func TestAlternatingWriters(t *testing.T) {
	dir := t.TempDir()
	ctx := context.Background()

	st1, err := Open(dir)
	if err != nil {
		t.Fatal(err)
	}

	defer func() { _ = st1.Close() }()

	st2, err := Open(dir)
	if err != nil {
		t.Fatal(err)
	}

	defer func() { _ = st2.Close() }()

	w, s := newBook(t, st1)
	rev := w.SavedRev()

	for i := 0; i < 10; i++ {
		st := st1
		if i%2 == 1 {
			st = st2
		}

		cell := workbook.ParseInput(strconv.Itoa(i))

		cmd := workbook.Command{Sheet: s.ID(), Edits: []workbook.CellEdit{
			{Pos: workbook.Pos{Row: i, Col: 0}, Cell: cell},
		}}

		next, err := st.SaveCells(ctx, w.ID(), s.ID(), cmd.Edits, rev, "alternate")
		if err != nil {
			t.Fatalf("save %d: %v", i, err)
		}

		rev = next
	}

	got, err := st1.LoadWorkbook(ctx, w.ID())
	if err != nil {
		t.Fatal(err)
	}

	if got.Rev() != 10 {
		t.Fatalf("rev = %d, want 10", got.Rev())
	}

	for i := 0; i < 10; i++ {
		want := strconv.Itoa(i)
		if v := got.Sheets()[0].Display(workbook.Pos{Row: i, Col: 0}); v != want {
			t.Fatalf("row %d = %q, want %q", i, v, want)
		}
	}
}
