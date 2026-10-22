package main

import (
	"context"
	"testing"
)

// TestWindowing renders 280 rows through a 900px viewport and checks the
// row window: few operations, full laid-out grid height, 36px rows, and a
// window that follows the scroll offset.
func TestWindowing(t *testing.T) {
	ctx := context.Background()
	app, err := newApp(280)
	if err != nil {
		t.Fatal(err)
	}
	if err := app.page.Redraw(ctx); err != nil {
		t.Fatal(err)
	}

	st := app.page.Probe().Stats()
	if st.Ops >= 800 {
		t.Fatalf("windowed ops = %d, want < 800", st.Ops)
	}

	boxes := app.page.Boxes()
	var gridH, rowH float64
	found := false
	for _, b := range boxes {
		switch b.ID {
		case "grid":
			gridH = b.H
		case "row-0":
			rowH, found = b.H, true
		}
	}
	if !found {
		t.Fatal("row-0 missing from window")
	}
	if rowH != 36 {
		t.Fatalf("row height = %v, want 36 (rowH constant)", rowH)
	}
	if gridH < 280*36 {
		t.Fatalf("grid height = %v, want >= %d (spacers)", gridH, 280*36)
	}

	app.page.Scroll().SetScrollOffset(0, 2000)
	if err := app.page.Redraw(ctx); err != nil {
		t.Fatal(err)
	}
	if app.winStart <= 0 {
		t.Fatalf("winStart = %d after scroll, want > 0", app.winStart)
	}
	if st2 := app.page.Probe().Stats(); st2.Ops >= 800 {
		t.Fatalf("scrolled ops = %d, want < 800", st2.Ops)
	}

	app.ScrollToRow(200)
	req, ok := app.page.Testing().TakeScroll()
	if !ok {
		t.Fatal("no scroll request from ScrollToRow")
	}
	want := int(app.gridTop()) + 200*36
	if req.Y != want {
		t.Fatalf("scroll Y = %d, want %d", req.Y, want)
	}
}
