package storage

import (
	"context"
	"os"
	"testing"

	"github.com/chinmay-sawant/ownframe/examples/spreadsheet/workbook"
)

func TestDeletedStorageReopensFresh(t *testing.T) {
	dir := t.TempDir()
	ctx := context.Background()

	st, err := Open(dir)
	if err != nil {
		t.Fatal(err)
	}

	if _, err := st.SeedWith(ctx, 1, workbook.SeedStress(50)); err != nil {
		t.Fatal(err)
	}

	_ = st.Close()

	if err := os.RemoveAll(dir); err != nil {
		t.Fatal(err)
	}

	st, err = Open(dir)
	if err != nil {
		t.Fatal(err)
	}

	defer func() { _ = st.Close() }()

	if v, err := st.SeedVersion(ctx); err != nil || v != 0 {
		t.Fatalf("SeedVersion = %d, %v, want a fresh database", v, err)
	}
}
