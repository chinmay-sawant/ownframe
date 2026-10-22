package main

import (
	"context"
	"testing"

	"github.com/chinmay-sawant/ownframe"
)

// TestStressOps redraws the dashboard headless and checks the node load.
func TestStressOps(t *testing.T) {
	ctx := context.Background()
	page, err := ownframe.NewWithOptions(ownframe.Config{
		Title: "Stress", HTML: buildHTML(), Width: 1280, Height: 900,
	}, ownframe.WithPerf(true))
	if err != nil {
		t.Fatal(err)
	}
	page.SetData(makeView(280))
	if err := page.Redraw(ctx); err != nil {
		t.Fatal(err)
	}
	d := page.Display()
	if d == nil {
		t.Fatal("no display after Redraw")
	}
	if got := len(d.Ops); got < 2000 {
		t.Fatalf("ops = %d, want > 2000", got)
	}
	st := page.Stats()
	if st.Ops < 2000 {
		t.Fatalf("stats ops = %d, want > 2000", st.Ops)
	}
	if st.Boxes < 1000 {
		t.Fatalf("boxes = %d, want > 1000", st.Boxes)
	}
	if st.Parses == 0 || st.Layouts == 0 || st.Redraws == 0 {
		t.Fatalf("stats not tracked: %+v", st)
	}
	if st.LastRedraw == 0 && st.LayoutTime == 0 {
		t.Fatalf("no perf timing: %+v", st)
	}
}
