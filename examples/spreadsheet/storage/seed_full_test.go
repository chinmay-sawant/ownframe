package storage

import (
	"context"
	"testing"
	"time"

	"github.com/chinmay-sawant/ownframe/examples/spreadsheet/workbook"
)

// TestSeedDummyFull seeds the shipped fixtures, including the 100,000-row
// stress workbook, and checks the second call is a no-op.
func TestSeedDummyFull(t *testing.T) {
	st, dir := openDir(t)
	ctx := context.Background()

	start := time.Now()

	wrote, err := st.SeedDummy(ctx)
	if err != nil || !wrote {
		t.Fatalf("SeedDummy wrote=%v err=%v", wrote, err)
	}

	t.Logf("full seed took %s", time.Since(start))

	if v, err := st.SeedVersion(ctx); err != nil || v != SeedVersion {
		t.Fatalf("SeedVersion = %d, %v", v, err)
	}

	page, err := st.ListWorkbooks(ctx, 0, 0)
	if err != nil {
		t.Fatal(err)
	}

	if len(page.Items) != 2 {
		t.Fatalf("workbooks = %d, want 2", len(page.Items))
	}

	var stress workbook.ID

	for _, it := range page.Items {
		if it.Name == "Stress sparse" {
			stress = it.ID
		}
	}

	if stress == 0 {
		t.Fatal("stress workbook missing")
	}

	start = time.Now()

	w, err := st.LoadWorkbook(ctx, stress)
	if err != nil {
		t.Fatal(err)
	}

	t.Logf("stress load took %s", time.Since(start))

	if v := w.Sheets()[0].Display(workbook.Pos{Row: 99999, Col: 0}); v != "99999" {
		t.Fatalf("stress A100000 = %q", v)
	}

	wrote, err = st.SeedDummy(ctx)
	if err != nil || wrote {
		t.Fatalf("second SeedDummy wrote=%v err=%v", wrote, err)
	}

	_ = st.Close()

	st, err = Open(dir)
	if err != nil {
		t.Fatal(err)
	}

	defer func() { _ = st.Close() }()

	page, err = st.ListWorkbooks(ctx, 0, 0)
	if err != nil || len(page.Items) != 2 {
		t.Fatalf("after reopen: %d workbooks, %v", len(page.Items), err)
	}
}

func TestJournalModes(t *testing.T) {
	st, _ := openDir(t)

	if mode := st.JournalMode(); mode != "wal" {
		t.Fatalf("file journal mode = %q, want wal", mode)
	}

	mem := openMemory(t)

	if mode := mem.JournalMode(); mode != "memory" {
		t.Fatalf("memory journal mode = %q", mode)
	}
}
