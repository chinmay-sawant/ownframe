package storage

import (
	"context"
	"testing"

	"github.com/chinmay-sawant/ownframe/examples/spreadsheet/workbook"
)

func TestListWorkbooksKeyset(t *testing.T) {
	st := openMemory(t)
	ctx := context.Background()

	for i := 0; i < 120; i++ {
		if _, err := st.CreateWorkbook(ctx, "Book"); err != nil {
			t.Fatal(err)
		}
	}

	var ids []workbook.ID

	cursor := workbook.ID(0)

	for page := 0; page < 3; page++ {
		p, err := st.ListWorkbooks(ctx, cursor, 0)
		if err != nil {
			t.Fatal(err)
		}

		want := 50
		if page == 2 {
			want = 20
		}

		if len(p.Items) != want {
			t.Fatalf("page %d = %d items, want %d", page, len(p.Items), want)
		}

		if p.More != (page < 2) {
			t.Fatalf("page %d More = %v", page, p.More)
		}

		for _, it := range p.Items {
			ids = append(ids, it.ID)
		}

		cursor = p.Next
	}

	if len(ids) != 120 {
		t.Fatalf("collected %d ids", len(ids))
	}

	for i := 1; i < len(ids); i++ {
		if ids[i] <= ids[i-1] {
			t.Fatalf("ids not ascending at %d: %v", i, ids[i-1:i+1])
		}
	}
}

func TestDeleteCascades(t *testing.T) {
	st, dir := openDir(t)
	w, s := newBook(t, st)

	edit(t, st, w, s, workbook.Pos{Row: 0, Col: 0}, "1")

	if err := st.DeleteWorkbook(context.Background(), w.ID()); err != nil {
		t.Fatal(err)
	}

	var cells int

	if err := raw(t, dir).QueryRowContext(t.Context(), `SELECT COUNT(*) FROM cells`).Scan(&cells); err != nil {
		t.Fatal(err)
	}

	if cells != 0 {
		t.Fatalf("cells after delete = %d, want foreign-key cleanup", cells)
	}

	var sheets int

	if err := raw(t, dir).QueryRowContext(t.Context(), `SELECT COUNT(*) FROM sheets`).Scan(&sheets); err != nil {
		t.Fatal(err)
	}

	if sheets != 0 {
		t.Fatalf("sheets after delete = %d", sheets)
	}
}
