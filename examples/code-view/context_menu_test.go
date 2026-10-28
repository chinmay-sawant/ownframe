package main

import (
	"context"
	"slices"
	"strings"
	"testing"

	"github.com/chinmay-sawant/ownframe"
)

func TestContextMenuTabRows(t *testing.T) {
	app := newTestApp(t)
	app.openFile("backend/probe.go")

	rows := app.ContextMenuFor("tab-0")
	if len(rows) != 5 {
		t.Fatalf("rows = %d, want 5", len(rows))
	}

	labels := []string{"Close", "Close Others", "Close All", "Close All to the Left", "Close All to the Right"}
	for i, label := range labels {
		if rows[i].Label != label {
			t.Errorf("row %d label = %q, want %q", i, rows[i].Label, label)
		}

		if !strings.HasPrefix(rows[i].ID, "code-view:") {
			t.Errorf("row %d id = %q, want the code-view prefix", i, rows[i].ID)
		}

		if !rows[i].Enabled && i != 3 {
			t.Errorf("row %d %q is disabled with two tabs", i, rows[i].Label)
		}
	}

	if rows := app.ContextMenuFor("tab-details"); len(rows) != 0 {
		t.Fatalf("tab-details rows = %d, want none", len(rows))
	}

	if rows := app.ContextMenuFor("editor"); len(rows) != 0 {
		t.Fatalf("editor rows = %d, want none", len(rows))
	}
}

func TestContextMenuCloseTabsOnEitherSide(t *testing.T) {
	for _, tc := range []struct {
		name       string
		group      int
		item       int
		active     int
		want       []string
		wantActive int
	}{
		{"left preserves active", 0, 3, 2, []string{"b", "c"}, 1},
		{"left closes active", 1, 3, 0, []string{"b", "c"}, 0},
		{"right preserves active", 1, 4, 0, []string{"a", "b"}, 0},
		{"right closes active", 0, 4, 2, []string{"a", "b"}, 1},
	} {
		t.Run(tc.name, func(t *testing.T) {
			app := newTestApp(t)
			app.tabs = []string{"a", "b", "c"}
			app.tabs2 = []string{"a", "b", "c"}
			app.split = true
			id := "tab-1"
			if tc.group == 1 {
				id = "tab2-1"
				app.active2 = tc.active
			} else {
				app.active = tc.active
			}
			rows := app.ContextMenuFor(id)
			if !rows[tc.item].Enabled {
				t.Fatalf("side action disabled: %+v", rows)
			}
			if err := app.MenuAction(context.Background(), rows[tc.item].ID); err != nil {
				t.Fatal(err)
			}
			got, active, other := app.tabs, app.active, app.tabs2
			if tc.group == 1 {
				got, active, other = app.tabs2, app.active2, app.tabs
			}
			if !slices.Equal(got, tc.want) || active != tc.wantActive || !slices.Equal(other, []string{"a", "b", "c"}) {
				t.Fatalf("tabs=%v active=%d other=%v", got, active, other)
			}
		})
	}
}

func TestContextMenuCloseOthersNeedsTwoTabs(t *testing.T) {
	app := newTestApp(t)

	rows := app.ContextMenuFor("tab-0")
	if len(rows) != 5 {
		t.Fatalf("rows = %d, want 5", len(rows))
	}

	if rows[1].Enabled {
		t.Fatal("Close Others is enabled with one tab")
	}
}

func TestContextMenuCloseAll(t *testing.T) {
	app := newTestApp(t)
	app.openFile("backend/probe.go")

	rows := app.ContextMenuFor("tab-0")
	if len(rows) != 5 {
		t.Fatalf("rows = %d, want 5", len(rows))
	}

	if err := app.MenuAction(context.Background(), rows[2].ID); err != nil {
		t.Fatalf("MenuAction: %v", err)
	}

	if len(app.tabs) != 0 {
		t.Fatalf("tabs = %v, want empty", app.tabs)
	}

	if app.active != 0 {
		t.Fatalf("active = %d, want 0", app.active)
	}

	if app.View().Tabs != nil && len(app.View().Tabs) != 0 {
		t.Fatalf("view tabs = %v, want empty", app.View().Tabs)
	}
}

func TestPageMenuClosesTheTab(t *testing.T) {
	app := newTestApp(t)
	app.openFile("backend/probe.go")
	ctx := context.Background()

	before := append([]string(nil), app.tabs...)
	rows := app.contextMenuForPage("tab-0")
	if len(rows) == 0 || len(before) == 0 {
		t.Fatal("tab has no menu rows")
	}

	if !app.presentMenu(40, 50, rows) || !app.View().MenuOpen {
		t.Fatal("page menu did not open")
	}

	if app.View().MenuX < 0 || app.View().MenuY < 0 {
		t.Fatalf("menu origin = %d,%d", app.View().MenuX, app.View().MenuY)
	}

	box, ok := boxByID(app.Page().Boxes(), "ctx-0")
	if !ok {
		t.Fatal("menu row was not painted")
	}

	if err := app.onClick(ctx, ownframe.Box{ID: box.ID, Action: box.Action}); err != nil {
		t.Fatal(err)
	}

	if app.menuOpen || app.View().MenuOpen {
		t.Fatal("menu stayed open after the row ran")
	}

	if len(before) == 0 || slices.Contains(app.tabs, before[0]) {
		t.Fatalf("tabs = %v, want %s closed", app.tabs, before)
	}
}

func TestPageMenuClickOutsideClosesIt(t *testing.T) {
	app := newTestApp(t)
	app.openFile("backend/probe.go")
	ctx := context.Background()
	app.scrollTo(0)
	if err := app.Page().Redraw(ctx); err != nil {
		t.Fatal(err)
	}

	rows := app.contextMenuForPage("tab-0")
	if !app.presentMenu(40, 50, rows) || !app.View().MenuOpen {
		t.Fatal("page menu did not open")
	}

	details, ok := boxByID(app.Page().Boxes(), "tab-details")
	if !ok {
		t.Fatal("missing details button")
	}

	if err := app.Page().Click(ctx, details.X+details.W/2, details.Y+details.H/2); err != nil {
		t.Fatal(err)
	}

	if app.menuOpen || app.View().MenuOpen {
		t.Fatal("menu stayed open after a click outside it")
	}

	if app.details {
		t.Fatal("outside click opened details")
	}
}

func TestPageMenuPressOutsideLeavesTheEditor(t *testing.T) {
	app := newTestApp(t)
	app.openFile("backend/probe.go")
	ctx := context.Background()
	app.scrollTo(0)
	if err := app.Page().Redraw(ctx); err != nil {
		t.Fatal(err)
	}

	rows := app.contextMenuForPage("tab-0")
	if !app.presentMenu(40, 50, rows) {
		t.Fatal("page menu did not open")
	}

	line, ok := boxByID(app.Page().Boxes(), "edit-line-0")
	if !ok {
		t.Fatal("missing editor line")
	}

	if err := app.Page().Edit().Press(ctx, line.X+line.W/2, line.Y+line.H/2); err != nil {
		t.Fatal(err)
	}

	if app.menuOpen || app.editing {
		t.Fatalf("menu=%v editing=%v", app.menuOpen, app.editing)
	}
}

// A centered label keeps its baseline in this band of the row. A label
// aligned to the top of the row sits above it.
const (
	menuLabelBaselineMin = 0.45
	menuLabelBaselineMax = 0.85
)

func TestPageMenuLabelIsVerticallyCentered(t *testing.T) {
	app := newTestApp(t)
	app.openFile("backend/probe.go")

	rows := app.contextMenuForPage("tab-0")
	if !app.presentMenu(40, 80, rows) {
		t.Fatal("page menu did not open")
	}

	box, ok := boxByID(app.Page().Boxes(), "ctx-0")
	if !ok {
		t.Fatal("menu row was not painted")
	}

	display := app.Page().Probe().Display()
	if display == nil || display.PointsPerPixel == 0 {
		t.Fatal("no display")
	}

	baseline, found := menuLabelBaseline(display, box, "Close")
	if !found {
		t.Fatal("Close label was not painted in the row")
	}

	top := box.Y + box.H*menuLabelBaselineMin
	bottom := box.Y + box.H*menuLabelBaselineMax
	if baseline < top || baseline > bottom {
		t.Fatalf("Close baseline %.1f, want between %.1f and %.1f", baseline, top, bottom)
	}
}

func menuLabelBaseline(display *ownframe.Display, box ownframe.Box, label string) (float64, bool) {
	for _, op := range display.Ops {
		if op.Text != label {
			continue
		}

		y := op.Y / display.PointsPerPixel
		if y < box.Y || y > box.Y+box.H {
			continue
		}

		return y, true
	}

	return 0, false
}
