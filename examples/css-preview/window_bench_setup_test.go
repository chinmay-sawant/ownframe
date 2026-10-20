package main

import (
	"context"
	"path/filepath"
	"testing"

	"github.com/chinmay-sawant/ownframe"
	"github.com/chinmay-sawant/ownframe/internal/previewwindow"
)

// windowBenchPage mirrors examples/css-preview/run.go: 960x860, perf on
// before the setup draws, then the catalog window installed and the
// devtools dock closed the way the walk hook closes it.
func windowBenchPage(tb testing.TB) (context.Context, *ownframe.Page, *view, *demo, *catalogMetrics) {
	tb.Helper()
	ctx := context.Background()
	v := &view{Started: map[string]bool{}}
	p, err := previewwindow.New(ctx, previewwindow.Config{
		FixtureDir: filepath.Join("testdata", "css-preview"),
		Data:       v, Width: 960, Height: 860})
	if err != nil {
		tb.Fatal(err)
	}
	p.SetPerf(true)
	p.SetDevTools(true)
	v.Catalog = catalogRows(nil)
	v.Rows = v.Catalog
	if err := previewwindow.Redraw(ctx, p, v); err != nil {
		tb.Fatal(err)
	}
	v.Catalog = catalogRows(p)
	v.Rows = v.Catalog
	v.Readout = describe(p)
	if err := previewwindow.Redraw(ctx, p, v); err != nil {
		tb.Fatal(err)
	}
	m, ok := measureCatalog(p)
	if !ok {
		tb.Fatal("could not measure the catalog")
	}
	d := newDemo(p, v)
	p.Handle(ownframe.Handlers{Click: d.click})
	if err := installCatalogWindow(ctx, p, v, d, m); err != nil {
		tb.Fatal(err)
	}
	p.SetDevTools(false)
	return ctx, p, v, d, m
}

// windowBenchReset returns the page to the top and clears started state.
func windowBenchReset(ctx context.Context, p *ownframe.Page, v *view, d *demo, tb testing.TB) {
	tb.Helper()
	v.Started = map[string]bool{}
	v.Selected = ""
	v.Readout = ""
	d.samples = map[string]*sample{}
	d.pendingKey = ""
	p.SetScrollOffset(0, 0)
	if p.StepScrollWindow() {
		if err := p.Redraw(ctx); err != nil {
			tb.Fatal(err)
		}
	}
	p.SetData(v)
	if err := p.Redraw(ctx); err != nil {
		tb.Fatal(err)
	}
}
