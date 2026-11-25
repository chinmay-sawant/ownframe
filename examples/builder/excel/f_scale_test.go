package main

import (
	"strconv"
	"strings"
	"testing"
	"time"
)

// fillFullSheet populates every one of the MaxRows x MaxCols cells with
// numbers and puts a running total in column A of row MaxRows-1, so a
// single-cell edit has a long recalculation chain behind it.
func fillFullSheet(sh *Sheet) {
	for r := 0; r < MaxRows; r++ {
		for c := 0; c < MaxCols; c++ {
			sh.SetRaw(r, c, strconv.Itoa(r*MaxCols+c))
		}
	}
	// B1 = A1; each row's column B chains to the cell above.
	for r := 1; r < MaxRows; r++ {
		sh.SetRaw(r, 1, "=B"+strconv.Itoa(r)+"+A"+strconv.Itoa(r+1))
	}
}

func newFullApp(t testing.TB) *app {
	if tt, ok := t.(*testing.T); ok {
		tt.Helper()
	}
	a, err := newApp()
	if err != nil {
		t.Fatal(err)
	}
	a.sheet().SetRaw(0, 1, "=A1")
	fillFullSheet(a.sheet())
	a.redraw()
	if a.lastErr != nil {
		t.Fatal(a.lastErr)
	}
	return a
}

// TestFullSheetBoundedByViewport pins EXL-72: a 1000 x 52 populated sheet
// mounts only the visible cells, and scrolling to the far corner keeps the
// painted node count and display list in the same order of magnitude.
func TestFullSheetBoundedByViewport(t *testing.T) {
	a := newFullApp(t)
	cells := func() int { return strings.Count(a.page.HTML(), `id="xl-gc-`) }
	start := cells()
	cw, ch := a.canvasSize()
	maxVisible := (cw/defaultColWidth + 2) * (ch/defaultRowHeight + 2)
	if start == 0 || start > maxVisible {
		t.Fatalf("mounted %d cells, want 1..%d for the viewport", start, maxVisible)
	}
	total := MaxRows * MaxCols
	if start*20 > total {
		t.Fatalf("%d of %d cells mounted: not virtualized", start, total)
	}
	boxes0 := len(a.page.Probe().Display().Boxes)

	for _, to := range [][2]int{{500, 20}, {MaxRows, MaxCols}, {0, 0}} {
		a.scrollRow, a.scrollCol = to[0], to[1]
		a.clampScroll()
		a.redraw()
		if n := cells(); n == 0 || n > maxVisible {
			t.Errorf("scroll to %v mounted %d cells, want 1..%d", to, n, maxVisible)
		}
		if n := len(a.page.Probe().Display().Boxes); n > boxes0*2 {
			t.Errorf("scroll to %v: %d display boxes vs %d at the top", to, n, boxes0)
		}
	}
}

// TestFullSheetScrollIsViewportBound compares one scroll step on the full
// sheet with one on the small sample sheet: cost follows the viewport, not
// the 52,000 populated cells. The bound is loose to stay stable on CI.
func TestFullSheetScrollIsViewportBound(t *testing.T) {
	if testing.Short() {
		t.Skip("populates 52,000 cells")
	}
	small, _ := newTestApp(t)
	full := newFullApp(t)
	step := func(a *app) time.Duration {
		best := time.Hour
		for i := 0; i < 5; i++ {
			a.scrollBy(1, 0)
			t0 := time.Now()
			a.redraw()
			best = min(best, time.Since(t0))
		}
		return best
	}
	s, f := step(small), step(full)
	t.Logf("scroll step: sample sheet %v, full sheet %v", s, f)
	if f > 16*s+50*time.Millisecond {
		t.Errorf("full-sheet scroll step %v is not O(visible) next to %v", f, s)
	}
	if f > 100*time.Millisecond {
		t.Errorf("full-sheet scroll step %v exceeds the 100ms sanity ceiling", f)
	}
}
