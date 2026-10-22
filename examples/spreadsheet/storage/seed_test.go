package storage

import (
	"context"
	"testing"

	"github.com/chinmay-sawant/ownframe/examples/spreadsheet/workbook"
)

func TestSeedIdempotent(t *testing.T) {
	st := openMemory(t)
	ctx := context.Background()

	wrote, err := st.SeedWith(ctx, 1, workbook.SeedStress(200))
	if err != nil || !wrote {
		t.Fatalf("first seed wrote=%v err=%v", wrote, err)
	}

	wrote, err = st.SeedWith(ctx, 1, workbook.SeedStress(200))
	if err != nil || wrote {
		t.Fatalf("second seed wrote=%v err=%v, want skipped", wrote, err)
	}

	page, err := st.ListWorkbooks(ctx, 0, 0)
	if err != nil {
		t.Fatal(err)
	}

	if len(page.Items) != 1 {
		t.Fatalf("workbooks = %d, want 1", len(page.Items))
	}

	if v, err := st.SeedVersion(ctx); err != nil || v != 1 {
		t.Fatalf("SeedVersion = %d, %v", v, err)
	}
}

func TestSeedPreservesUserChanges(t *testing.T) {
	dir := t.TempDir()
	ctx := context.Background()

	st, err := Open(dir)
	if err != nil {
		t.Fatal(err)
	}

	if _, err := st.SeedWith(ctx, 1, workbook.SeedStress(50)); err != nil {
		t.Fatal(err)
	}

	page, err := st.ListWorkbooks(ctx, 0, 0)
	if err != nil {
		t.Fatal(err)
	}

	id := page.Items[0].ID

	w, err := st.LoadWorkbook(ctx, id)
	if err != nil {
		t.Fatal(err)
	}

	s := w.Sheets()[0]

	edit(t, st, w, s, workbook.Pos{Row: 0, Col: 0}, "user change")

	_ = st.Close()

	st, err = Open(dir)
	if err != nil {
		t.Fatal(err)
	}

	defer func() { _ = st.Close() }()

	wrote, err := st.SeedWith(ctx, 1, workbook.SeedStress(50))
	if err != nil || wrote {
		t.Fatalf("reseed wrote=%v err=%v", wrote, err)
	}

	got, err := st.LoadWorkbook(ctx, id)
	if err != nil {
		t.Fatal(err)
	}

	if v := got.Sheets()[0].Display(workbook.Pos{Row: 0, Col: 0}); v != "user change" {
		t.Fatalf("A1 = %q, user change was overwritten", v)
	}
}
