package main

import (
	"context"
	"fmt"
	"testing"
)

// TestTabWindowing overflows eight long-named tabs at a wide and a narrow
// window: the strip shows full-width tabs with scroll arrows, hidden tabs
// do not render, and the arrows move the window.
func TestTabWindowing(t *testing.T) {
	dir := t.TempDir()
	names := []string{
		"android-build-complete.log",
		"android-build-final.log",
		"dino-android-build.log",
		"fallback-final-tests.log",
		"final-tests-after-atomic.log",
		"cancel-tests.log",
		"gpu-complete.log",
		"benchmarks.log",
	}

	app, err := New(dir)
	if err != nil {
		t.Fatal(err)
	}

	ctx := context.Background()
	for _, n := range names {
		writeFile(t, dir, n, "hello\nworld\n")
		app.openFile(n)
	}

	// Wide: opening ends on the last tab, so the window sits at the tail
	// with the prev arrow showing.
	app.Page().SetSize(1908, 999)
	if err := app.Page().Redraw(ctx); err != nil {
		t.Fatal(err)
	}

	tail := app.View()
	if len(tail.Tabs) < 3 || len(tail.Tabs) >= len(names) {
		t.Fatalf("wide: %d tabs render, want a middle window", len(tail.Tabs))
	}

	if !tail.TabsPrev {
		t.Fatal("wide: prev arrow missing at the tail window")
	}

	if tail.TabsNext {
		t.Fatal("wide: next arrow shows with the last tab visible")
	}

	if tail.Tabs[len(tail.Tabs)-1].Index != len(names)-1 {
		t.Fatal("wide: active last tab not in the window")
	}

	assertTabWindow(t, app, tail, names)

	// Back to the first tab: the window snaps home, prev hides, next shows.
	app.selectTab(0)
	if err := app.Page().Redraw(ctx); err != nil {
		t.Fatal(err)
	}

	view := app.View()
	if len(view.Tabs) == 0 || view.Tabs[0].Index != 0 {
		t.Fatalf("select first: window starts at %+v, want tab 0", view.Tabs)
	}

	if view.TabsPrev {
		t.Fatal("select first: prev arrow shows at the first window")
	}

	if !view.TabsNext {
		t.Fatal("select first: next arrow missing with hidden tabs")
	}

	assertTabWindow(t, app, view, names)

	// The strip tools stay pinned to the strip's right edge.
	strip, ok := boxByID(app.Page().Boxes(), "tabstrip")
	if !ok {
		t.Fatal("missing tabstrip")
	}

	for _, id := range []string{"split-toggle", "tab-details"} {
		tool, ok := boxByID(app.Page().Boxes(), id)
		if !ok {
			t.Fatalf("missing %s", id)
		}

		if gap := strip.X + strip.W - (tool.X + tool.W); gap < -1 || gap > 33 {
			t.Fatalf("%s ends %.1f px before the strip edge", id, gap)
		}
	}

	// The next arrow sticks to the tools' left, not to the last tab.
	next, ok := boxByID(app.Page().Boxes(), "tabs-next")
	if !ok {
		t.Fatal("missing tabs-next")
	}

	split, ok := boxByID(app.Page().Boxes(), "split-toggle")
	if !ok {
		t.Fatal("missing split-toggle")
	}

	if gap := split.X - (next.X + next.W); gap < 0 || gap > 5 {
		t.Fatalf("tabs-next ends %.1f px before split-toggle, want 0-5 px", gap)
	}

	// Narrow: fewer tabs fit, but every rendered tab still shows its full
	// name without overlapping its neighbour.
	app.Page().SetSize(1100, 700)
	if err := app.Page().Redraw(ctx); err != nil {
		t.Fatal(err)
	}

	narrow := app.View()
	if len(narrow.Tabs) >= len(view.Tabs) {
		t.Fatalf("narrow: %d tabs render, want fewer than wide %d", len(narrow.Tabs), len(view.Tabs))
	}

	assertTabWindow(t, app, narrow, names)

	// The next arrow moves the window forward and reveals the prev arrow.
	clickID(t, app, "tabs-next")
	moved := app.View()
	if len(moved.Tabs) == 0 || moved.Tabs[0].Index != view.Tabs[0].Index+1 {
		t.Fatalf("next: window starts at %+v, want one tab later", moved.Tabs)
	}

	if !moved.TabsPrev {
		t.Fatal("next: prev arrow missing after scrolling forward")
	}

	assertTabWindow(t, app, moved, names)

	// Selecting the last tab pulls it into view.
	app.selectTab(len(names) - 1)
	if err := app.Page().Redraw(ctx); err != nil {
		t.Fatal(err)
	}

	last := app.View()
	found := false
	for _, row := range last.Tabs {
		if row.Index == len(names)-1 {
			found = true
		}
	}
	if !found {
		t.Fatal("select: last tab not in the window")
	}

	if last.TabsNext {
		t.Fatal("select: next arrow shows with the last tab visible")
	}

	assertTabWindow(t, app, last, names)
}

// assertTabWindow requires every rendered tab to sit apart from its
// neighbour, paint its full name, and keep its close button at the tab's
// right edge. Hidden tabs must not paint at all.
func assertTabWindow(t *testing.T, app *App, view View, names []string) {
	t.Helper()

	boxes := app.Page().Boxes()
	display := app.Page().Probe().Display()
	if display == nil {
		t.Fatal("page fell back to a bitmap")
	}

	visible := make(map[int]bool, len(view.Tabs))
	var prevX, prevW float64

	for i, row := range view.Tabs {
		visible[row.Index] = true

		box, ok := boxByID(boxes, fmt.Sprintf("tab-%d", row.Index))
		if !ok {
			t.Fatalf("missing tab-%d", row.Index)
		}

		if i > 0 && box.X < prevX+prevW-1 {
			t.Fatalf("tab-%d at x=%.1f overlaps previous tab ending at %.1f",
				row.Index, box.X, prevX+prevW)
		}
		prevX, prevW = box.X, box.W

		found := false
		for _, op := range display.Ops {
			if op.Text == names[row.Index] {
				found = true
				break
			}
		}
		if !found {
			t.Errorf("full name %q not painted", names[row.Index])
		}

		closeBox, ok := boxByID(boxes, fmt.Sprintf("tab-close-%d", row.Index))
		if !ok {
			t.Fatalf("missing tab-close-%d", row.Index)
		}

		if gap := box.X + box.W - (closeBox.X + closeBox.W); gap < 0 || gap > 8 {
			t.Errorf("tab-close-%d ends %.1f px before the tab edge, want 0-8 px",
				row.Index, gap)
		}
	}

	for i, name := range names {
		if visible[i] {
			continue
		}

		// The active tab's name still paints in the crumbs bar below the
		// strip; only its strip card must stay hidden.
		if i != view.ActiveTab {
			for _, op := range display.Ops {
				if op.Text == name {
					t.Errorf("hidden tab %q paints", name)
					break
				}
			}
		}

		if _, ok := boxByID(boxes, fmt.Sprintf("tab-%d", i)); ok {
			t.Errorf("hidden tab-%d has a box", i)
		}
	}
}
