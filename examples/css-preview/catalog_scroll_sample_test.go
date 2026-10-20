package main

import (
	"context"
	"testing"
)

// TestScrollKeepsTheRunningSample checks that a vertical scroll does not
// stop or reset an animation that is already running. The sample keeps its
// origin, so it resumes where it was when the row comes back.
func TestScrollKeepsTheRunningSample(t *testing.T) {
	ctx := context.Background()
	p, d, v := fixturePage(t)
	m, ok := measureCatalog(p)
	if !ok {
		t.Fatal("could not measure the catalog")
	}
	if err := installCatalogWindow(ctx, p, v, d, m); err != nil {
		t.Fatal(err)
	}

	startCatalog(t, p, d, "-webkit-animation")
	s := d.samples[catalogSampleKey("-webkit-animation")]
	if s == nil || !s.active {
		t.Fatal("the sample did not start")
	}
	origin := s.origin

	p.SetScrollOffset(0, m.rowTop[4]-200)
	if p.StepScrollWindow() {
		if err := p.Redraw(ctx); err != nil {
			t.Fatal(err)
		}
	}
	if err := d.tick(ctx); err != nil {
		t.Fatal(err)
	}
	if s.origin != origin {
		t.Fatal("the scroll reset the sample origin")
	}
	if !s.active {
		t.Fatal("the scroll stopped the running sample")
	}

	p.SetScrollOffset(0, 0)
	if p.StepScrollWindow() {
		if err := p.Redraw(ctx); err != nil {
			t.Fatal(err)
		}
	}
	if err := d.tick(ctx); err != nil {
		t.Fatal(err)
	}
	if s.detached {
		t.Fatal("the sample did not attach again")
	}
}
