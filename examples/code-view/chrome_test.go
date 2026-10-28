package main

import (
	"context"
	"testing"
)

func TestChromeControlsFollowFlag(t *testing.T) {
	app := newTestApp(t)
	ctx := context.Background()
	ids := []string{"win-min", "win-max", "win-close"}

	if _, ok := boxByID(app.Page().Boxes(), "win-min"); ok {
		t.Fatal("window controls painted without the chrome flag")
	}

	app.chrome = true
	app.refresh()

	if err := app.Page().Redraw(ctx); err != nil {
		t.Fatal(err)
	}

	view := app.View()
	if !view.Chrome {
		t.Fatal("chrome flag did not reach the view")
	}

	if view.ChromeMax {
		t.Fatal("headless window reports maximized")
	}

	for _, id := range ids {
		box, ok := boxByID(app.Page().Boxes(), id)
		if !ok {
			t.Fatalf("missing %s with chrome on", id)
		}

		if box.H <= 0 || box.W <= 0 {
			t.Fatalf("%s has no clickable area", id)
		}
	}

	// Minimize and maximize are no-ops without a desktop run loop; they
	// must not error. The close control is left unclicked: it arms the
	// real run loop exit.
	clickID(t, app, "win-max")
	clickID(t, app, "win-min")
}
