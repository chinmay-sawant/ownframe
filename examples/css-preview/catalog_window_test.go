package main

import (
	"context"
	"fmt"
	"testing"

	"github.com/chinmay-sawant/ownframe"
)

// TestCatalogWindowKeepsFullGeometry checks that the row window keeps the
// catalog grid and metadata section at their full-page positions.
func TestCatalogWindowKeepsFullGeometry(t *testing.T) {
	ctx := context.Background()
	p, d, v := fixturePage(t)
	m, ok := measureCatalog(p)
	if !ok {
		t.Fatal("could not measure the full catalog")
	}
	fullGrid, ok := findBox(p.Boxes(), "catalog-grid")
	if !ok {
		t.Fatal("missing full catalog grid")
	}
	fullMeta, ok := findBox(p.Boxes(), "metadata")
	if !ok {
		t.Fatal("missing full metadata section")
	}

	if err := installCatalogWindow(ctx, p, v, d, m); err != nil {
		t.Fatal(err)
	}
	if len(v.Rows) >= len(v.Catalog) {
		t.Fatalf("window has %d rows, want fewer than %d", len(v.Rows), len(v.Catalog))
	}
	checkCatalogGeometry(t, p, fullGrid, fullMeta, "top")

	p.SetScrollOffset(0, m.rowTop[10]-200)
	if err := p.Redraw(ctx); err != nil {
		t.Fatal(err)
	}
	checkCatalogGeometry(t, p, fullGrid, fullMeta, "middle")
	if _, ok := findBox(p.Boxes(), fmt.Sprintf("catalog-row-%02d", 21)); !ok {
		t.Fatal("row 21 is missing from the middle window")
	}

	p.SetScrollOffset(0, m.contentH)
	if err := p.Redraw(ctx); err != nil {
		t.Fatal(err)
	}
	checkCatalogGeometry(t, p, fullGrid, fullMeta, "bottom")
}

func checkCatalogGeometry(t *testing.T, p *ownframe.Page, fullGrid, fullMeta ownframe.Box, where string) {
	t.Helper()
	grid, ok := findBox(p.Boxes(), "catalog-grid")
	if !ok {
		t.Fatalf("%s: missing catalog grid", where)
	}
	meta, ok := findBox(p.Boxes(), "metadata")
	if !ok {
		t.Fatalf("%s: missing metadata section", where)
	}
	if diff := grid.H - fullGrid.H; diff > 3 || diff < -3 {
		t.Fatalf("%s: grid height %.1f, want %.1f", where, grid.H, fullGrid.H)
	}
	if diff := meta.Y - fullMeta.Y; diff > 3 || diff < -3 {
		t.Fatalf("%s: metadata y %.1f, want %.1f", where, meta.Y, fullMeta.Y)
	}
}
