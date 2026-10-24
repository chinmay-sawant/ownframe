package main

import (
	"context"
	"testing"
	"time"
)

// TestScrollStartsTheRowThatComesIntoView checks that a row that scrolls
// into view with its started class binds its preview, and that the automatic
// driver then keeps it moving.
func TestScrollStartsTheRowThatComesIntoView(t *testing.T) {
	ctx := context.Background()
	p, d, v := fixturePage(t)
	m, ok := measureCatalog(p)
	if !ok {
		t.Fatal("could not measure the catalog")
	}
	if err := installCatalogWindow(ctx, p, v, d, m); err != nil {
		t.Fatal(err)
	}
	clickID(t, p, playAllID)
	tickPage(t, p)

	late := v.Catalog[29]
	if _, found := findBox(p.Boxes(), late.TargetID); found {
		t.Fatal("the late row was already in the window")
	}
	scrollTo(t, ctx, p, m, m.rowTop[14]-200)
	time.Sleep(5 * time.Millisecond)
	tickPage(t, p)
	if _, found := findBox(p.Boxes(), late.TargetID); !found {
		t.Fatal("the late row did not scroll into the window")
	}
	if got := p.Probe().Motion().Bound; got == 0 {
		t.Fatalf("bound players = %d after the scroll", got)
	}
	if got := xformX(targetOp(t, p, late.TargetID), p.Probe().Display().PointsPerPixel); got == 0 {
		t.Fatalf("%s did not bind after it scrolled into view", late.Name)
	}
}
