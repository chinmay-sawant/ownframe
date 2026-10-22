package storage

import (
	"context"
	"testing"

	"github.com/chinmay-sawant/ownframe/examples/spreadsheet/workbook"
)

func BenchmarkLoadRange(b *testing.B) {
	ctx := context.Background()

	st, err := Open(Memory)
	if err != nil {
		b.Fatal(err)
	}

	defer func() { _ = st.Close() }()

	if _, err := st.SeedWith(ctx, 1, workbook.SeedStress(10000)); err != nil {
		b.Fatal(err)
	}

	page, err := st.ListWorkbooks(ctx, 0, 0)
	if err != nil {
		b.Fatal(err)
	}

	w, err := st.LoadWorkbook(ctx, page.Items[0].ID)
	if err != nil {
		b.Fatal(err)
	}

	sheet := w.Sheets()[0].ID()
	rect := workbook.Rect{Min: workbook.Pos{Row: 9000, Col: 0}, Max: workbook.Pos{Row: 9099, Col: 3}}

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		if _, err := st.LoadRange(ctx, sheet, rect, 500); err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkSaveCells(b *testing.B) {
	ctx := context.Background()

	st, err := Open(Memory)
	if err != nil {
		b.Fatal(err)
	}

	defer func() { _ = st.Close() }()

	w, err := st.CreateWorkbook(ctx, "bench")
	if err != nil {
		b.Fatal(err)
	}

	sheet := w.Sheets()[0].ID()
	rev := w.SavedRev()

	edits := make([]workbook.CellEdit, 256)

	for i := range edits {
		edits[i] = workbook.CellEdit{
			Pos:  workbook.Pos{Row: i, Col: 0},
			Cell: workbook.ParseInput("1"),
		}
	}

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		next, err := st.SaveCells(ctx, w.ID(), sheet, edits, rev, "bench")
		if err != nil {
			b.Fatal(err)
		}

		rev = next
	}
}
