package main

import (
	"context"
	"testing"
)

// TestSideWheelScrollsTabList wheels over the pinned tab list in a short
// window and checks the last tab button comes into the visible list.
func TestSideWheelScrollsTabList(t *testing.T) {
	app, err := NewApp(catalogPath())
	if err != nil {
		t.Fatal(err)
	}

	for _, tab := range app.tabs {
		t.Run(tab.ID, func(t *testing.T) {
			t.Parallel()
			sideWheelCheck(t, app, tab.ID)
		})
	}
}

func sideWheelCheck(t *testing.T, app *App, id string) {
	t.Helper()

	ctx := context.Background()

	page, err := app.NewPage(ctx, id, 1000, 500)
	if err != nil {
		t.Fatal(err)
	}

	find := func(id string) (x, y, w, h float64) {
		for _, b := range page.Boxes() {
			if b.ID == id {
				return b.X, b.Y, b.W, b.H
			}
		}
		t.Fatalf("no box %q", id)

		return
	}

	sx, sy, sw, sh := find("side")
	_, _, _, ih := find("side-in")
	t.Logf("side %v,%v %vx%v inner h %v", sx, sy, sw, sh, ih)

	if ih <= sh {
		t.Skip("tab list fits")
	}

	last := "tab-btn-" + app.tabs[len(app.tabs)-1].ID
	_, ly, _, lh := find(last)

	if ly+lh <= sy+sh {
		t.Fatalf("last tab already visible")
	}

	for i := 0; i < 200; i++ {
		if _, err = sideWheel(page)(ctx, sx+sw/2, sy+sh/2, 0, -1); err != nil {
			t.Fatal(err)
		}
	}

	_, ly, _, lh = find(last)
	_, sy, _, sh = find("side")

	if ly+lh > sy+sh+0.5 {
		t.Errorf("last tab bottom %v below list bottom %v", ly+lh, sy+sh)
	}
}
