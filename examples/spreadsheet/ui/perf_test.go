package ui

import (
	"context"
	"fmt"
	"testing"
)

// TestPerfBreakdown prints the stage timings and operation mix for one
// scroll step. Run with -v to read the numbers.
func TestPerfBreakdown(t *testing.T) {
	fake := newFake()
	app, err := New(Options{Backend: fake, Width: 1200, Height: 800, Perf: true})
	if err != nil {
		t.Fatal(err)
	}

	defer func() { _ = app.Close() }()
	settle(t, app)
	ctx := context.Background()
	app.page.SetScrollOffset(0, 480)
	app.page.Testing().StepScrollWindow()
	if err := app.page.Redraw(ctx); err != nil {
		t.Fatal(err)
	}

	s := app.page.Probe().Stats()
	d := app.page.Probe().Display()
	kinds := map[string]int{}
	for i := range d.Ops {
		kinds[fmt.Sprint(d.Ops[i].Kind)]++
	}

	t.Logf("ops=%d kinds=%v template=%v layout=%v display=%v draw=%v redraw=%v",
		len(d.Ops), kinds, s.LastTemplate, s.LayoutTime, s.DisplayListTime, s.LastDraw, s.LastRedraw)
}
