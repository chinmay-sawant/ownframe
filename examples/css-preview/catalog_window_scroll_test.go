package main

import (
	"context"
	"testing"
	"time"

	"github.com/chinmay-sawant/ownframe"
)

// catalogScrollTo moves the window to the row that starts at the given top.
func catalogScrollTo(ctx context.Context, t *testing.T, p *ownframe.Page, top int) {
	t.Helper()
	p.SetScrollOffset(0, top)
	if !p.Testing().StepScrollWindow() {
		t.Fatal("the row window did not move")
	}
	if err := p.Redraw(ctx); err != nil {
		t.Fatal(err)
	}
}

// TestScrollQueuesTheRowThatComesIntoView checks that a row that scrolls
// into view starts its preview and that the queue drains again. A queue that
// never drains keeps the window awake forever.
func TestScrollQueuesTheRowThatComesIntoView(t *testing.T) {
	ctx := context.Background()
	p, d, v := fixturePage(t)
	m, ok := measureCatalog(p)
	if !ok {
		t.Fatal("could not measure the catalog")
	}
	if err := installCatalogWindow(ctx, p, v, d, m); err != nil {
		t.Fatal(err)
	}
	if err := d.playAll(ctx); err != nil {
		t.Fatal(err)
	}
	if err := d.tick(ctx); err != nil {
		t.Fatal(err)
	}
	if len(d.pending) != 0 {
		t.Fatalf("pending = %d, want an empty queue", len(d.pending))
	}
	catalogScrollTo(ctx, t, p, m.rowTop[16])
	if len(d.pending) == 0 {
		t.Fatal("the row that came into view was not queued")
	}
	d.pendingAt = time.Now().Add(-time.Second)
	if err := d.tick(ctx); err != nil {
		t.Fatal(err)
	}
	if len(d.pending) != 0 {
		t.Fatalf("pending = %d, want an empty queue", len(d.pending))
	}
	started := 0
	for _, s := range d.samples {
		if s.active {
			started++
		}
	}
	if started == 0 {
		t.Fatal("no preview started after the scroll")
	}
}
