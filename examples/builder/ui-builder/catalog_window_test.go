package main

import (
	"context"
	"strings"
	"testing"

	"github.com/chinmay-sawant/ownframe/ui"
)

func catalogAbsInt(value int) int {
	if value < 0 {
		return -value
	}
	return value
}

func TestCatalogWindowRangeMath(t *testing.T) {
	t.Parallel()
	const (
		total    = 39
		header   = 96
		pitch    = 300
		viewH    = 476
		overscan = catalogWindowOverscanRows
	)
	tests := []struct {
		name      string
		offsetY   int
		viewH     int
		pitch     int
		overscan  int
		wantStart int
		wantEnd   int
	}{
		{"top", 0, viewH, pitch, overscan, 0, 4},
		{"scrolled one pitch", 300, viewH, pitch, overscan, 0, 5},
		{"middle", 3000, viewH, pitch, overscan, 7, 14},
		{"bottom clamp", 20000, viewH, pitch, overscan, 39, 39},
		{"negative offset clamps", -1000, viewH, pitch, overscan, 0, 1},
		{"zero pitch renders full", 3000, viewH, 0, overscan, 0, total},
		{"zero viewport renders full", 3000, 0, pitch, overscan, 0, total},
		{"negative overscan treated as zero", 3000, viewH, pitch, -5, 9, 12},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			start, end := catalogWindowRange(total, tc.offsetY, tc.viewH, header, tc.pitch, tc.overscan)
			if start != tc.wantStart || end != tc.wantEnd {
				t.Fatalf("range(%d) = (%d,%d), want (%d,%d)", tc.offsetY, start, end, tc.wantStart, tc.wantEnd)
			}
			if start < 0 || end < start || end > total {
				t.Fatalf("range(%d) = (%d,%d) out of bounds", tc.offsetY, start, end)
			}
		})
	}
	if start, end := catalogWindowRange(0, 100, viewH, header, pitch, overscan); start != 0 || end != 0 {
		t.Fatalf("empty catalog range = (%d,%d), want (0,0)", start, end)
	}
}

// TestCatalogWindowSpacerMath pins the height-preservation identity: the top
// stub plus the bottom stub plus the visible block estimate equals the full
// rows height, so updateCenterScrollMetrics keeps reporting the full content
// height from the content box. The laid-out container additionally pays one
// gap per rendered stub, which the overscan band absorbs.
func TestCatalogWindowSpacerMath(t *testing.T) {
	t.Parallel()
	const (
		totalRows = 39
		pitch     = 100
		gap       = catalogWindowGapPx
		rowsH     = totalRows*pitch - gap
	)
	visibleBlock := func(start, end int) int { return (end-start)*pitch - gap }
	for _, tc := range []struct {
		name       string
		start, end int
		wantTop    int
		wantBottom int
	}{
		{"interior", 10, 14, 1000, 2500},
		{"top edge omits stub", 0, 4, 0, 3500},
		{"bottom edge omits stub", 35, 39, 3500, 0},
	} {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			top, bottom := catalogWindowSpacers(tc.start, tc.end, totalRows, pitch, gap, rowsH)
			if top != tc.wantTop || bottom != tc.wantBottom {
				t.Fatalf("spacers(%d,%d) = (%d,%d), want (%d,%d)", tc.start, tc.end, top, bottom, tc.wantTop, tc.wantBottom)
			}
			if got := top + bottom + visibleBlock(tc.start, tc.end); got != rowsH {
				t.Fatalf("spacers(%d,%d) preserve %d, want rows height %d", tc.start, tc.end, got, rowsH)
			}
		})
	}
	if top, bottom := catalogWindowSpacers(0, 0, 0, pitch, gap, 0); top != 0 || bottom != 0 {
		t.Fatalf("empty spacers = (%d,%d), want (0,0)", top, bottom)
	}
	if top, bottom := catalogWindowSpacers(10, 14, totalRows, pitch, gap, 0); top+bottom+visibleBlock(10, 14) != rowsH {
		t.Fatalf("fallback rows height spacers preserve %d, want %d", top+bottom+visibleBlock(10, 14), rowsH)
	}
}

func TestCatalogWindowEntryMapping(t *testing.T) {
	t.Parallel()
	if got := catalogRowCount(0); got != 0 {
		t.Fatalf("rows(0) = %d, want 0", got)
	}
	if got := catalogRowCount(1); got != 1 {
		t.Fatalf("rows(1) = %d, want 1", got)
	}
	if got := catalogRowCount(78); got != 39 {
		t.Fatalf("rows(78) = %d, want 39", got)
	}
	for _, tc := range []struct {
		name                string
		total, start, end   int
		wantFirst, wantLast int
	}{
		{"full catalog", 78, 0, 39, 0, 78},
		{"odd tail", 5, 2, 3, 4, 5},
		{"clamped past end", 78, 40, 50, 78, 78},
		{"empty", 0, 0, 0, 0, 0},
	} {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			first, last := catalogWindowEntries(tc.total, tc.start, tc.end)
			if first != tc.wantFirst || last != tc.wantLast {
				t.Fatalf("entries(%d,%d,%d) = (%d,%d), want (%d,%d)",
					tc.total, tc.start, tc.end, first, last, tc.wantFirst, tc.wantLast)
			}
		})
	}
}

func renderCatalogRows(t *testing.T, rows []*ui.Element) string {
	t.Helper()
	html, err := ui.Render(ui.Div().Children(rows...))
	if err != nil {
		t.Fatal(err)
	}
	return html
}

func delimited(name string) string { return ">" + name + "<" }

func TestCatalogWindowedRowsSlice(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	a, err := newBuilderApp()
	if err != nil {
		t.Fatal(err)
	}
	a.page.SetSize(900, 610)
	a.tab = "catalog"
	if err := a.refresh(ctx); err != nil {
		t.Fatal(err)
	}
	if err := a.refreshOnResize(ctx); err != nil {
		t.Fatal(err)
	}
	entries := catalogEntries()
	if len(entries) != 78 {
		t.Fatalf("entries=%d, want 78", len(entries))
	}

	t.Run("top window", func(t *testing.T) {
		plan := a.catalogWindowPlan()
		if !plan.windowed {
			t.Fatal("compact catalog should window")
		}
		if plan.startRow != 0 {
			t.Fatalf("top start=%d, want 0", plan.startRow)
		}
		rows := a.catalogWindowedRows()
		wantRows := plan.endRow - plan.startRow + 1 // no top stub, one bottom stub
		if len(rows) != wantRows {
			t.Fatalf("built %d rows, want %d", len(rows), wantRows)
		}
		html := renderCatalogRows(t, rows)
		if !strings.Contains(html, delimited(entries[0].name)) {
			t.Fatal("top window missing first entry")
		}
		if strings.Contains(html, delimited(entries[len(entries)-1].name)) {
			t.Fatal("top window built the last entry")
		}
		if strings.Contains(html, catalogWindowTopSpacerID) {
			t.Fatal("top window should omit the top stub")
		}
		if !strings.Contains(html, catalogWindowBottomSpacerID) {
			t.Fatal("top window missing the bottom stub")
		}
	})

	t.Run("scrolled window", func(t *testing.T) {
		max := a.center.Metrics().MaxOffset()
		if max == 0 {
			t.Fatal("compact catalog should overflow")
		}
		if !a.center.ScrollTo(max / 2) {
			t.Fatal("scroll to middle failed")
		}
		plan := a.catalogWindowPlan()
		if !plan.windowed || plan.startRow <= 0 {
			t.Fatalf("scrolled plan = %+v, want interior window", plan)
		}
		first, last := catalogWindowEntries(len(entries), plan.startRow, plan.endRow)
		rows := a.catalogWindowedRows()
		html := renderCatalogRows(t, rows)
		for _, name := range []string{entries[first].name, entries[last-1].name} {
			if !strings.Contains(html, delimited(name)) {
				t.Fatalf("scrolled window missing visible entry %q (plan %+v)", name, plan)
			}
		}
		if strings.Contains(html, delimited(entries[0].name)) {
			t.Fatal("scrolled window built the first entry")
		}
		if plan.endRow < plan.totalRows && strings.Contains(html, delimited(entries[len(entries)-1].name)) {
			t.Fatal("scrolled window built the last entry")
		}
	})

	t.Run("other tab renders full", func(t *testing.T) {
		a.tab = "widgets"
		defer func() { a.tab = "catalog" }()
		if plan := a.catalogWindowPlan(); plan.windowed {
			t.Fatalf("non-catalog plan = %+v, want full", plan)
		}
		html := renderCatalogRows(t, a.catalogWindowedRows())
		if strings.Contains(html, catalogWindowTopSpacerID) || strings.Contains(html, catalogWindowBottomSpacerID) {
			t.Fatal("full render should carry no spacer stubs")
		}
	})
}

func TestCatalogWindowingPlumbing(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	a, err := newBuilderApp()
	if err != nil {
		t.Fatal(err)
	}
	a.page.SetSize(900, 610)
	a.tab = "catalog"
	if err := a.refresh(ctx); err != nil {
		t.Fatal(err)
	}
	if err := a.refreshOnResize(ctx); err != nil {
		t.Fatal(err)
	}
	if !a.page.Scroll().Windowing() {
		t.Fatal("refresh tick should install the catalog window callback")
	}
	viewport := a.centerViewport()
	if !a.catalogOnScroll(0, int(viewport.H)) {
		t.Fatal("overflowing catalog at top should imply a partial window")
	}
	a.tab = "widgets"
	if a.catalogOnScroll(0, int(viewport.H)) {
		t.Fatal("non-catalog tab should never imply a window")
	}
	a.tab = "catalog"

	initial := a.center.Metrics()
	if initial.MaxOffset() == 0 {
		t.Fatal("compact catalog should overflow")
	}
	for range 3 {
		if _, err := a.handleWheel(ctx, viewport.X+viewport.W/2, viewport.Y+20, 0, -1); err != nil {
			t.Fatal(err)
		}
		if err := a.refreshOnResize(ctx); err != nil {
			t.Fatal(err)
		}
		// Windowed renders estimate the visible band with the average row
		// pitch, so the total may breathe by a fraction of one band while
		// the scrollbar range stays put. Bound the drift instead of
		// demanding an exact match: two pitches covers the estimation
		// error without hiding a runaway.
		_, pitchPx, _, ok := a.catalogWindowGeometry()
		if !ok {
			t.Fatal("windowed catalog should keep measurable geometry")
		}
		if got := a.center.Metrics(); catalogAbsInt(got.Content-initial.Content) > 2*pitchPx {
			t.Fatalf("windowing plumbing moved content height: before=%+v after=%+v pitch=%d", initial, got, pitchPx)
		}
	}
}
