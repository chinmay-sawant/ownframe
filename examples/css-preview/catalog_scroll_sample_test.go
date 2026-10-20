package main

import (
	"context"
	"testing"

	"github.com/chinmay-sawant/ownframe"
)

// TestScrollKeepsEveryRunningSample checks that a vertical scroll never stops
// or resets a running animation, whatever kind it is. A sample whose row
// leaves the window keeps its origin and its destination, and it attaches
// again when the row comes back.
func TestScrollKeepsEveryRunningSample(t *testing.T) {
	names := []string{
		"-webkit-animation", "-webkit-animation-iteration-count", "animation",
		"animation-trigger", "transition", "transition-property", "view-transition-name",
	}
	for _, name := range names {
		t.Run(name, func(t *testing.T) {
			ctx := context.Background()
			p, d, v := fixturePage(t)
			m, ok := measureCatalog(p)
			if !ok {
				t.Fatal("could not measure the catalog")
			}
			if err := installCatalogWindow(ctx, p, v, d, m); err != nil {
				t.Fatal(err)
			}
			row := -1
			for i, c := range v.Catalog {
				if c.Name == name {
					row = i
				}
			}
			if row < 0 {
				t.Fatalf("missing catalog row %q", name)
			}
			scrollTo(t, ctx, p, m, m.rowTop[rowAt(m, row/2)]-200)

			startCatalog(t, p, d, name)
			s := d.samples[catalogSampleKey(name)]
			if s == nil || !s.active {
				t.Fatal("the sample did not start")
			}
			origin, to := s.origin, s.to

			scrollTo(t, ctx, p, m, m.rowTop[rowAt(m, row/2+4)]-200)
			if err := d.tick(ctx); err != nil {
				t.Fatal(err)
			}
			if !s.active {
				t.Fatal("the scroll stopped the running sample")
			}
			if s.origin != origin {
				t.Fatal("the scroll reset the sample origin")
			}
			if !samePaints(to, s.to) {
				t.Fatal("the scroll changed the sample destination")
			}

			scrollTo(t, ctx, p, m, m.rowTop[rowAt(m, row/2)]-200)
			if err := d.tick(ctx); err != nil {
				t.Fatal(err)
			}
			if s.detached {
				t.Fatal("the sample did not attach again")
			}
			if !s.active {
				t.Fatal("the sample stopped after it came back")
			}
		})
	}
}

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
