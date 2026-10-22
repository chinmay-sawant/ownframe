package main

import (
	"context"
	"testing"

	"github.com/chinmay-sawant/ownframe"
)

func scrollTo(t *testing.T, ctx context.Context, p *ownframe.Page, m *catalogMetrics, y int) {
	t.Helper()
	p.SetScrollOffset(0, y)
	if p.StepScrollWindow() {
		if err := p.Redraw(ctx); err != nil {
			t.Fatal(err)
		}
	}
}

// rowAt clamps a row index to the measured rows.
func rowAt(m *catalogMetrics, i int) int {
	if i < 0 {
		return 0
	}
	if i >= len(m.rowTop) {
		return len(m.rowTop) - 1
	}
	return i
}

func samePaints(a, b []paint) bool {
	if len(a) != len(b) {
		return false
	}
	for i := range a {
		if a[i] != b[i] {
			return false
		}
	}
	return true
}
