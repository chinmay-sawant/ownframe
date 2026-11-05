package main

import (
	"context"
	"testing"

	"github.com/chinmay-sawant/ownframe"
)

func TestPinnedChromeLayer(t *testing.T) {
	a, err := newApp()
	if err != nil {
		t.Fatal(err)
	}
	ctx := context.Background()
	if err := a.page.Redraw(ctx); err != nil {
		t.Fatal(err)
	}
	// TestPinnedChromeLayer guards the overlay hit-testing contract: the fixed
	// topbar and the search backdrop must emit Fixed ops so the window paints
	// (and hit-tests) them above the scrolling body. Without this, overlay
	// clicks lose to the body boxes beneath them.
	countFixed := func() int {
		n := 0
		for _, op := range a.page.Probe().Display().Ops {
			if op.Fixed {
				n++
			}
		}
		return n
	}
	closed := countFixed()
	t.Logf("closed fixed ops: %d", closed)
	if closed == 0 {
		t.Fatalf("fixed topbar produces no Fixed ops; overlay clicks will lose to body")
	}
	if err := a.handleClick(ctx, ownframe.Box{Action: "search:open"}); err != nil {
		t.Fatal(err)
	}
	if err := a.page.Redraw(ctx); err != nil {
		t.Fatal(err)
	}
	open := countFixed()
	t.Logf("open fixed ops: %d", open)
	if open <= closed {
		t.Fatalf("search backdrop adds no Fixed ops; click-away will miss")
	}
}
