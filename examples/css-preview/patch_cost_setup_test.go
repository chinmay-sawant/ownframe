package main

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	"github.com/chinmay-sawant/ownframe"
)

// patchCostWidth and patchCostHeight are the desktop window size the CSS
// preview runs at.
const (
	patchCostWidth  = 1908
	patchCostHeight = 999
)

// patchCostPage builds the CSS preview fixture at the desktop window size.
// The whole catalog is in the page, and the engine keeps its layout state so
// a single element can be patched.
func patchCostPage(tb testing.TB) (context.Context, *ownframe.Page, *view) {
	tb.Helper()
	ctx := context.Background()
	dir := filepath.Join("testdata", "css-preview")
	source, err := os.ReadFile(filepath.Join(dir, "index.html")) //nolint:gosec // reason: fixed repo-relative testdata path joined to a constant dir.
	if err != nil {
		tb.Fatal(err)
	}
	theme, err := os.ReadFile(filepath.Join(dir, "styles.css")) //nolint:gosec // reason: fixed repo-relative testdata path joined to a constant dir.
	if err != nil {
		tb.Fatal(err)
	}
	v := &view{Started: map[string]bool{}}
	p, err := ownframe.New(ownframe.Config{HTML: string(source), Theme: string(theme),
		Width: patchCostWidth, Height: patchCostHeight, RetainState: true})
	if err != nil {
		tb.Fatal(err)
	}
	p.Window().SetPerf(true)
	p.SetData(v)
	if err := p.Redraw(ctx); err != nil {
		tb.Fatal(err)
	}
	v.Catalog = catalogRows(nil)
	v.Rows = v.Catalog
	p.SetData(v)
	if err := p.Redraw(ctx); err != nil {
		tb.Fatal(err)
	}
	v.Catalog = catalogRows(p)
	v.Rows = v.Catalog
	v.Readout = describe(p)
	p.SetData(v)
	if err := p.Redraw(ctx); err != nil {
		tb.Fatal(err)
	}
	return ctx, p, v
}
