package main

import (
	"context"
	"testing"
)

func TestCullWindowExample(t *testing.T) {
	t.Parallel()

	ctx := context.Background()

	full, err := newCullWindowPage(false, false, true)
	if err != nil {
		t.Fatalf("newCullWindowPage(false) failed: %v", err)
	}
	if err := full.Redraw(ctx); err != nil {
		t.Fatalf("full Redraw failed: %v", err)
	}

	culled, err := newCullWindowPage(true, false, true)
	if err != nil {
		t.Fatalf("newCullWindowPage(true) failed: %v", err)
	}
	if err := culled.Redraw(ctx); err != nil {
		t.Fatalf("culled Redraw failed: %v", err)
	}

	fullDisp := full.Probe().Display()
	culledDisp := culled.Probe().Display()

	if culledDisp.Height != fullDisp.Height {
		t.Fatalf("culled Height = %d, want %d", culledDisp.Height, fullDisp.Height)
	}
	if len(culledDisp.Ops) >= len(fullDisp.Ops)/4 {
		t.Fatalf("culled ops = %d, full ops = %d; expected viewport culling", len(culledDisp.Ops), len(fullDisp.Ops))
	}

	// Scroll deep into the 1,000-row list (~row 450 at y=20000) and verify
	// the overscan window shifts to emit rows in that band.
	culled.Scroll().SetScrollOffset(0, 20000)
	if err := culled.Redraw(ctx); err != nil {
		t.Fatalf("scrolled culled Redraw failed: %v", err)
	}

	foundMid := false
	for _, op := range culled.Probe().Display().Ops {
		if op.Text == "Record 455" {
			foundMid = true
			break
		}
	}
	if !foundMid {
		t.Fatal("expected Record 455 in Display.Ops after scrolling to y=20000")
	}
}
