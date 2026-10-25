package main

import (
	"context"
	"testing"

	"github.com/chinmay-sawant/ownframe"
)

// windowBenchPage mirrors examples/css-preview/run.go: 960x860, perf on
// before the setup draws, then the catalog window installed and the
// devtools dock closed the way the walk hook closes it.
func windowBenchPage(tb testing.TB) (context.Context, *ownframe.Page, *view, *demo, *catalogMetrics) {
	return windowBenchPageSize(tb, 960, 860)
}

func windowBenchPageSize(tb testing.TB, width, height int) (context.Context, *ownframe.Page, *view, *demo, *catalogMetrics) {
	tb.Helper()
	ctx := context.Background()
	v := newView()
	p, err := openCatalog(ctx, catalogOpen{Data: v, Width: width, Height: height})
	if err != nil {
		tb.Fatal(err)
	}
	p.Window().SetPerf(true)
	p.Window().SetDevTools(true)
	v.Catalog = catalogRows(nil)
	v.Rows = v.Catalog
	if err := redrawCatalog(ctx, p, v); err != nil {
		tb.Fatal(err)
	}
	v.Catalog = catalogRows(p)
	v.Rows = v.Catalog
	v.Readout = describe(p)
	if err := redrawCatalog(ctx, p, v); err != nil {
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
	p.Window().SetDevTools(false)
	return ctx, p, v, d, m
}

// windowBenchReset returns the page to the top and clears started state.
func windowBenchReset(ctx context.Context, p *ownframe.Page, v *view, d *demo, tb testing.TB) {
	tb.Helper()
	v.Started = map[string]bool{}
	v.Restart = map[string]bool{}
	v.Restarting = map[string]bool{}
	v.Selected = ""
	v.Readout = ""
	d.manual = map[string]*manualSample{}
	p.Scroll().SetScrollOffset(0, 0)
	if p.Testing().StepScrollWindow() {
		if err := p.Redraw(ctx); err != nil {
			tb.Fatal(err)
		}
	}
	v.applyRows()
	p.SetData(v)
	if err := p.Redraw(ctx); err != nil {
		tb.Fatal(err)
	}
}
