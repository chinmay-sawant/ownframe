package main

import (
	"context"
	"testing"
	"time"

	"github.com/chinmay-sawant/ownframe"
)

// windowBenchScroll applies the scroll offset the way the window's scroll
// sync does: run the row window callback and redraw only when it changes.
func windowBenchScroll(ctx context.Context, p *ownframe.Page, tb testing.TB) {
	tb.Helper()
	if p.Testing().StepScrollWindow() {
		if err := p.Redraw(ctx); err != nil {
			tb.Fatal(err)
		}
	}
}

// windowBenchPass walks all 43 Start buttons. It mirrors the walk hook:
// scroll the row under a fixed line, drop the previous sample, then click.
func windowBenchPass(ctx context.Context, p *ownframe.Page, v *view,
	d *demo, m *catalogMetrics, tb testing.TB,
) []time.Duration {
	tb.Helper()
	windowBenchReset(ctx, p, v, d, tb)
	_, viewH := p.Size()
	out := make([]time.Duration, 0, len(v.Catalog))
	for i, row := range v.Catalog {
		y := m.rowTop[i/2] - 200
		if maxY := m.contentH - viewH; y > maxY {
			y = maxY
		}
		if y < 0 {
			y = 0
		}
		p.SetScrollOffset(0, y)
		windowBenchScroll(ctx, p, tb)
		d.samples = map[string]*sample{}
		d.pending = nil
		out = append(out, windowBenchClick(ctx, p, tb, row.Button))
	}
	return out
}
