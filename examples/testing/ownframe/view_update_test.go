package main

import (
	"context"
	"testing"

	"github.com/chinmay-sawant/ownframe/pagetest"
)

func TestInvoiceViewUpdatePatchesSelection(t *testing.T) {
	ctx := context.Background()
	a, err := newInvoiceApp()
	if err != nil {
		t.Fatal(err)
	}
	a.patch = false
	if err := a.Page().Redraw(ctx); err != nil {
		t.Fatal(err)
	}
	before := a.Page().Probe().Stats()
	if err := pagetest.ClickAction(ctx, a.Page(), "row:501"); err != nil {
		t.Fatal(err)
	}
	after := a.Page().Probe().Stats()
	t.Logf("view parses %d->%d cascades %d->%d layouts %d->%d",
		before.Parses, after.Parses,
		before.Cascades, after.Cascades,
		before.Layouts, after.Layouts)
	if after.Parses != before.Parses || after.Cascades != before.Cascades {
		t.Fatalf("view update fell back to a full render")
	}
}
