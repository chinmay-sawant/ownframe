package app

import (
	"context"
	"testing"
	"time"

	"github.com/chinmay-sawant/ownframe"
)

// navFill returns the background fill colour of one sidebar nav item. The box
// owns its subtree's operation ids, so the first fill in that set is the row
// background.
func navFill(t *testing.T, app *App, id string) [3]float64 {
	t.Helper()

	d := app.Page().Probe().Display()
	if d == nil {
		t.Fatal("the dashboard has no display list")
	}

	box, ok := findBox(app.Boxes(), id)
	if !ok {
		t.Fatalf("no %s box", id)
	}

	for _, owned := range box.OperationIDs {
		for i := range d.Ops {
			op := &d.Ops[i]
			if op.ID == owned && op.Kind == ownframe.DisplayOpFillRect {
				return [3]float64{op.R, op.G, op.B}
			}
		}
	}

	t.Fatalf("%s owns no background fill", id)

	return [3]float64{}
}

// strictlyBetween reports that every channel of mid sits inside the two
// endpoint colours.
func strictlyBetween(mid, a, b [3]float64) bool {
	for i := range len(mid) {
		lo, hi := min(a[i], b[i]), max(a[i], b[i])
		if mid[i] <= lo || mid[i] >= hi {
			return false
		}
	}

	return true
}

// TestSidebarBackgroundTransitionsFromCSS checks the nav highlight is a real
// transition: the click handler changes ActivePage, the template swaps the `on`
// class, and the automatic runtime moves the fill between the two colours. The
// example registers no tick and calls no motion package.
func TestSidebarBackgroundTransitionsFromCSS(t *testing.T) {
	const (
		probeDelay = 40 * time.Millisecond
		runLength  = 120 * time.Millisecond
	)

	app, err := New()
	if err != nil {
		t.Fatalf("New: %v", err)
	}

	ctx := context.Background()
	if err := app.Page().Redraw(ctx); err != nil {
		t.Fatalf("Redraw: %v", err)
	}

	inactive := navFill(t, app, "nav-insights")
	active := navFill(t, app, "nav-dictation")
	if inactive == active {
		t.Fatalf("both nav states paint the same colour %v", inactive)
	}

	// The sidebar starts on dictation. Alternate the target so a machine that
	// stalls past one run can try the reverse class change.
	targets := []string{"nav-insights", "nav-dictation"}
	for attempt := range 3 {
		target := targets[attempt%len(targets)]

		box, ok := findBox(app.Boxes(), target)
		if !ok {
			t.Fatalf("no %s box", target)
		}

		start := time.Now()
		clickBox(t, app, box)

		stats := app.Page().Probe().Motion()
		if stats.TransitionsStarted < 2 {
			t.Fatalf("after clicking %s: transitions started = %d, want 2",
				target, stats.TransitionsStarted)
		}
		if stats.TransitionError != "" {
			t.Fatalf("transition bind failed: %s", stats.TransitionError)
		}
		if app.Page().Probe().Ticking() {
			t.Fatal("the dashboard registered a tick callback")
		}

		time.Sleep(probeDelay)
		if err := app.Page().Testing().Tick(ctx); err != nil {
			t.Fatalf("Tick: %v", err)
		}

		got := navFill(t, app, target)
		if strictlyBetween(got, inactive, active) {
			return
		}
		if elapsed := time.Since(start); elapsed < runLength {
			t.Fatalf("after %v: %s paints %v, want a value between %v and %v",
				elapsed, target, got, inactive, active)
		}
		// The tick landed after the run, so the paint is the endpoint. Reverse
		// the class change and try again.
	}

	t.Fatal("three class changes never sampled the transition mid-run")
}
