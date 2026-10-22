package depthalert

import (
	"context"
	"image"
	"testing"
)

func TestTickUpdatesOnlySummaryValues(t *testing.T) {
	a, err := New()
	if err != nil {
		t.Fatal(err)
	}
	a.Page().Probe().TakeDirtyRects()
	before := a.Page().Probe().Generation()
	for _, box := range a.Page().Boxes() {
		if box.ID == "dashboard" && box.H > 280 {
			t.Fatalf("dashboard height = %.1f, exceeds the non-scroll panel height", box.H)
		}
	}
	for range 3 {
		a.ObserveDepth(2200, 60, true)
	}
	if err := a.Tick(context.Background()); err != nil {
		t.Fatal(err)
	}
	if got := a.Page().Probe().Generation(); got != before {
		t.Fatalf("generation = %d, want retained paint update at %d", got, before)
	}
	if got := a.ops.near.Text; got != "2.2 m" {
		t.Fatalf("distance operation = %q, want updated summary value", got)
	}
	if got := a.ops.coverage.Text; got != "60%" {
		t.Fatalf("coverage operation = %q, want updated summary value", got)
	}
	rects, ok := a.Page().Probe().TakeDirtyRects()
	if !ok {
		t.Fatal("summary value changes did not report damage")
	}
	for _, rect := range rects {
		if rect == image.Rect(0, 0, 360, 320) || rect.Dx() >= 360 {
			t.Fatalf("summary update damaged the full page: %v", rect)
		}
	}
}
