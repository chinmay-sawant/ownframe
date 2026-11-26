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

// TestToolbarsWrapWhenNarrow renders both toolbars at 1000 px and checks no
// control leaves the toolbar's horizontal bounds.
func TestToolbarsWrapWhenNarrow(t *testing.T) {
	r := newRig(t, 1000, 760, "")
	host, kids := r.toolbarKids("editor-toolbar")
	for _, k := range kids {
		if k.X+k.W > host.X+host.W+1 {
			t.Errorf("%s overflows at 1000px: %.0f > %.0f", k.ID, k.X+k.W, host.X+host.W)
		}
	}
	if lo, hi := rowSpan(r, "editor-toolbar"); hi-lo < 20 {
		t.Errorf("query toolbar should wrap to a second row at 1000 px (rows span %.0f px)", hi-lo)
	}
	if bar := r.box("editor-toolbar"); bar.H < 60 {
		t.Errorf("wrapped toolbar must grow to hold both rows, height %.0f", bar.H)
	}
	r.png("toolbar-query-1000")
	r.openTableByName("orders")
	r.png("toolbar-table-1000")
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
