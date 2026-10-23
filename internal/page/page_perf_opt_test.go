package page_test

import (
	"context"
	"testing"
)

// Perf stays off unless requested: an untracked Redraw leaves the opt-in
// Stats fields at zero, so end users pay nothing by default.
func TestPerfOffByDefault(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	p := newCachePage(t, `<p>hi</p>`)
	if p.Perf() {
		t.Fatal("perf must default to off")
	}
	if err := p.Redraw(ctx); err != nil {
		t.Fatal(err)
	}

	s := p.Stats()
	if s.LastTemplate != 0 || s.StyleTime != 0 || s.LayoutTime != 0 || s.PaintTime != 0 {
		t.Fatalf("stages tracked while off: %+v", s)
	}
	if s.AllocFrame != 0 || s.DirtyOps != 0 || s.DirtyRegions != 0 || s.ChangedOps != 0 {
		t.Fatalf("counters tracked while off: %+v", s)
	}
	if s.Redraws != 1 {
		t.Fatalf("redraws = %d, want 1", s.Redraws)
	}
}
