package ui

import (
	"testing"
)

// toolbarKids lists the laid-out boxes whose rectangle sits inside the toolbar
// with the given id and that carry an id of their own.
func (r *rig) toolbarKids(bar string) (host rigBox, kids []rigBox) {
	r.t.Helper()
	b := r.box(bar)
	host = rigBox{b.ID, b.X, b.Y, b.W, b.H}
	for _, c := range r.a.Page().Boxes() {
		if c.ID == "" || c.ID == bar {
			continue
		}
		if c.X >= b.X-1 && c.Y >= b.Y-1 && c.X < b.X+b.W && c.Y+c.H <= b.Y+b.H+1 && c.W > 0 && c.H > 0 && c.ID[:min(3, len(c.ID))] == "ed-" {
			kids = append(kids, rigBox{c.ID, c.X, c.Y, c.W, c.H})
		}
	}
	return host, kids
}

type rigBox struct {
	ID         string
	X, Y, W, H float64
}

// TestToolbarsFitAt1280 checks that no query or table toolbar control is cut
// off at the right edge of its toolbar when the window is 1280 px wide.
func TestToolbarsFitAt1280(t *testing.T) {
	r := newRig(t, 1280, 800, "")
	check := func(name, bar string) {
		host, kids := r.toolbarKids(bar)
		if len(kids) == 0 {
			t.Fatalf("%s: no controls found", name)
		}
		for _, k := range kids {
			if k.X+k.W > host.X+host.W+1 {
				t.Errorf("%s: %s overflows the toolbar: right edge %.0f > %.0f", name, k.ID, k.X+k.W, host.X+host.W)
			}
		}
	}
	check("query toolbar", "editor-toolbar")
	if lo, hi := rowSpan(r, "editor-toolbar"); hi-lo > 10 {
		t.Errorf("query toolbar wraps at 1280 px (rows span %.0f px)", hi-lo)
	}
	r.png("toolbar-query-1280")
	r.openTableByName("orders")
	r.png("toolbar-table-1280")
}

// TestToolbarOverflowWhenNarrow renders both toolbars at 1000 px: the low
// priority buttons collapse into a "..." menu instead of wrapping, and every
// action stays reachable through it.
func TestToolbarOverflowWhenNarrow(t *testing.T) {
	r := newRig(t, 1000, 760, "")
	host, kids := r.toolbarKids("editor-toolbar")
	for _, k := range kids {
		if k.X+k.W > host.X+host.W+1 {
			t.Errorf("%s overflows at 1000px: %.0f > %.0f", k.ID, k.X+k.W, host.X+host.W)
		}
	}
	if lo, hi := rowSpan(r, "editor-toolbar"); hi-lo > 10 {
		t.Errorf("query toolbar must not wrap at 1000 px (rows span %.0f px)", hi-lo)
	}
	if !r.has("ov-more-ed") || r.has("ed-snippet") || !r.has("ed-run") {
		t.Fatal("Save Snippet should have collapsed into the overflow menu, Run stays")
	}
	r.tap("ov-more-ed")
	for _, id := range []string{"ov-item-ed-snippet", "ov-item-ed-format", "ov-item-ed-wrap"} {
		if !r.has(id) {
			t.Fatalf("overflow menu lacks %s", id)
		}
	}
	r.png("toolbar-query-1000-menu")
	r.tap("ov-item-ed-snippet")
	if r.a.menuOpen != "" || r.has("ov-menu-ed") {
		t.Fatal("choosing an item closes the menu")
	}
	r.tap("ov-more-ed")
	r.key("escape")
	if r.has("ov-menu-ed") {
		t.Fatal("Escape closes the menu")
	}
	r.tap("ov-more-ed")
	r.tap("sb-title")
	if r.has("ov-menu-ed") {
		t.Fatal("a press elsewhere closes the menu")
	}
	r.openTableByName("orders")
	if !r.has("ov-more-tbl") {
		t.Fatal("the table toolbar collapses Import/Export at 1000 px")
	}
	r.png("toolbar-table-1000")
}

// TestToolbarNoOverflowWhenWide keeps every button visible when there is room.
func TestToolbarNoOverflowWhenWide(t *testing.T) {
	r := newRig(t, 1600, 900, "")
	for _, id := range []string{"ed-run", "ed-run-sel", "ed-explain", "ed-analyze", "ed-format", "ed-snippet"} {
		if !r.has(id) {
			t.Fatalf("%s should show at 1600 px", id)
		}
	}
}

// rowSpan is the vertical spread of the toolbar's controls: near zero on one
// row, a control height or more once it wraps.
func rowSpan(r *rig, bar string) (lo, hi float64) {
	_, kids := r.toolbarKids(bar)
	lo, hi = 1e9, -1e9
	for _, k := range kids {
		lo, hi = min(lo, k.Y), max(hi, k.Y)
	}
	return lo, hi
}
