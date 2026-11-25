package main

import (
	"context"
	"testing"
)

// TestFillHandleDragAutoFills drags the fill handle two rows down and expects
// the numeric series to continue.
func TestFillHandleDragAutoFills(t *testing.T) {
	a, _ := newTestApp(t)
	sh := a.sheet()
	sh.SetRaw(20, 12, "1")
	sh.SetRaw(21, 12, "2")
	a.selectRange(Range{Min: CellRef{Row: 20, Col: 12}, Max: CellRef{Row: 21, Col: 12}})
	if err := a.page.Redraw(context.Background()); err != nil {
		t.Fatal(err)
	}
	hb := mustBox(t, a, "xl-gov-handle")
	hx, hy := float64(hb.X)+float64(hb.W)/2, float64(hb.Y)+float64(hb.H)/2
	ctx := context.Background()
	if _, err := a.handlePress(ctx, hb, hx, hy); err != nil {
		t.Fatal(err)
	}
	if err := a.handleDragMove(ctx, 0, float64(a.zs(sh.RowHeight(22))+a.zs(sh.RowHeight(23)))); err != nil {
		t.Fatal(err)
	}
	if err := a.handleDragEnd(ctx); err != nil {
		t.Fatal(err)
	}
	if got := sh.Display(23, 12); got != "4" {
		t.Fatalf("M24 = %q, want 4 (series 1,2,3,4)", got)
	}
}
