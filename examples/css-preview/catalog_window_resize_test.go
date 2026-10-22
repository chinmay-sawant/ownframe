package main

import (
	"context"
	"testing"
	"time"
)

// TestResizeMeasuresTheCatalogAgain checks that a resize measures the
// catalog again. The geometry was measured at the old size, so the scroll
// range kept the old rows and the last row could not be reached.
func TestResizeMeasuresTheCatalogAgain(t *testing.T) {
	ctx := context.Background()
	p, d, v := fixturePage(t)
	m, ok := measureCatalog(p)
	if !ok {
		t.Fatal("could not measure the catalog")
	}
	if err := installCatalogWindow(ctx, p, v, d, m); err != nil {
		t.Fatal(err)
	}
	before := m.contentH
	p.SetSize(600, 900)
	if err := p.Redraw(ctx); err != nil {
		t.Fatal(err)
	}
	if !d.remeasure {
		t.Fatal("the resize was not noticed")
	}
	d.resizeAt = time.Now().Add(-time.Second)
	if err := d.tick(ctx); err != nil {
		t.Fatal(err)
	}
	after := d.window.metrics.contentH
	if after <= before {
		t.Fatalf("content height = %d, want more than %d", after, before)
	}
}
