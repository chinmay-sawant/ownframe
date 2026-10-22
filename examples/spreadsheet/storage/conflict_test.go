package storage

import (
	"context"
	"errors"
	"testing"

	"github.com/chinmay-sawant/ownframe/examples/spreadsheet/workbook"
)

func TestRevisionConflict(t *testing.T) {
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

	w1, s1 := newBook(t, st1)

	w2, err := st2.LoadWorkbook(ctx, w1.ID())
	if err != nil {
		t.Fatal(err)
	}

	s2 := w2.Sheets()[0]

	edit(t, st1, w1, s1, workbook.Pos{Row: 0, Col: 0}, "window one")

	cmd := workbook.Command{Sheet: s2.ID(), Edits: []workbook.CellEdit{
		{Pos: workbook.Pos{Row: 0, Col: 1}, Cell: workbook.ParseInput("window two")},
	}}

	if _, err := w2.Apply(cmd); err != nil {
		t.Fatal(err)
	}

	_, err = st2.SaveCells(ctx, w2.ID(), s2.ID(), cmd.Edits, w2.SavedRev(), "window two")
	if !errors.Is(err, ErrConflict) {
		t.Fatalf("second window save = %v, want ErrConflict", err)
	}

	w2, err = st2.LoadWorkbook(ctx, w1.ID())
	if err != nil {
		t.Fatal(err)
	}

	if v := w2.Sheets()[0].Display(workbook.Pos{Row: 0, Col: 1}); v != "" {
		t.Fatalf("conflicting save wrote %q", v)
	}

	cmd = workbook.Command{Sheet: w2.Sheets()[0].ID(), Edits: []workbook.CellEdit{
		{Pos: workbook.Pos{Row: 0, Col: 1}, Cell: workbook.ParseInput("window two")},
	}}

	if _, err := w2.Apply(cmd); err != nil {
		t.Fatal(err)
	}

	if _, err := st2.SaveCells(ctx, w2.ID(), w2.Sheets()[0].ID(), cmd.Edits, w2.SavedRev(), "retry"); err != nil {
		t.Fatalf("retry after reload: %v", err)
	}
}

func TestSaveUnknownWorkbook(t *testing.T) {
	st := openMemory(t)
	ctx := context.Background()

	cmd := []workbook.CellEdit{{Pos: workbook.Pos{Row: 0, Col: 0}, Cell: workbook.ParseInput("1")}}

	if _, err := st.SaveCells(ctx, 99, 1, cmd, 0, "test"); !errors.Is(err, ErrNotFound) {
		t.Fatalf("err = %v, want ErrNotFound", err)
	}
}
