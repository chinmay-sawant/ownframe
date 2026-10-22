package storage

import (
	"context"
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestDefaultDir(t *testing.T) {
	dir, err := DefaultDir()
	if err != nil {
		t.Skipf("no user config dir: %v", err)
	}

	if !strings.HasSuffix(dir, filepath.Join("ownframe", "spreadsheet")) {
		t.Fatalf("DefaultDir = %q", dir)
	}
}

func TestOpenOddPath(t *testing.T) {
	dir := filepath.Join(t.TempDir(), "spa ce#q?ü")
	ctx := context.Background()

	st, err := Open(dir)
	if err != nil {
		t.Fatal(err)
	}

	defer func() { _ = st.Close() }()

	if err := st.SetPref(ctx, "theme", "dark"); err != nil {
		t.Fatal(err)
	}

	v, found, err := st.GetPref(ctx, "theme")
	if err != nil || !found || v != "dark" {
		t.Fatalf("pref = %q %v %v", v, found, err)
	}

	if _, err := os.Stat(filepath.Join(dir, File)); err != nil {
		t.Fatal(err)
	}
}

func TestMemoryIsolation(t *testing.T) {
	a := openMemory(t)
	b := openMemory(t)
	ctx := context.Background()

	if _, err := a.CreateWorkbook(ctx, "only in a"); err != nil {
		t.Fatal(err)
	}

	page, err := b.ListWorkbooks(ctx, 0, 0)
	if err != nil {
		t.Fatal(err)
	}

	if len(page.Items) != 0 {
		t.Fatalf("memory stores share data: %d workbooks", len(page.Items))
	}

	page, err = a.ListWorkbooks(ctx, 0, 0)
	if err != nil || len(page.Items) != 1 {
		t.Fatalf("a lists %d workbooks, %v", len(page.Items), err)
	}
}

func TestClosedStoreRejectsCalls(t *testing.T) {
	st := openMemory(t)

	if err := st.Close(); err != nil {
		t.Fatal(err)
	}

	if err := st.Close(); err != nil {
		t.Fatal("second close:", err)
	}

	if _, err := st.CreateWorkbook(context.Background(), "x"); !errors.Is(err, ErrClosed) {
		t.Fatalf("call after close = %v, want ErrClosed", err)
	}
}
