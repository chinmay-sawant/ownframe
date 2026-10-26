package page_test

import (
	"context"
	"testing"

	"github.com/chinmay-sawant/ownframe/internal/page"
)

// First Redraw fills every stage and dirties the frame; TakeDirty keeps
// the reported counters.
func TestPerfFirstRedraw(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	p := newCachePage(t, `<p>hi</p>`)
	p.SetPerf(true)
	if err := p.Redraw(ctx); err != nil {
		t.Fatal(err)
	}

	s := p.Stats()
	if s.LastTemplate < 0 || s.StyleTime < 0 || s.LayoutTime < 0 || s.PaintTime < 0 {
		t.Fatalf("stages %v %v %v %v", s.LastTemplate, s.StyleTime, s.LayoutTime, s.PaintTime)
	}
	if s.LastRedraw <= 0 || s.AllocFrame == 0 {
		t.Fatalf("redraw %v alloc %d", s.LastRedraw, s.AllocFrame)
	}
	if s.DirtyRegions != 1 || s.DirtyOps <= 0 || s.ChangedOps <= 0 {
		t.Fatalf("dirty %d/%d changed %d", s.DirtyOps, s.DirtyRegions, s.ChangedOps)
	}

	p.TakeDirty()
	if n := p.Stats(); n.DirtyRegions != 1 || n.ChangedOps != s.ChangedOps {
		t.Fatalf("take cleared %d/%d", n.DirtyRegions, n.ChangedOps)
	}
}

// An edit reports a dirty region and a resize reports the whole frame.
func TestPerfDirtyCounts(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	p, err := page.New(page.Config{HTML: `<p>{{.V}}</p>`, Width: 320, Height: 200, Perf: true})
	if err != nil {
		t.Fatal(err)
	}

	redraw := func(v string) {
		t.Helper()
		p.SetData(struct{ V string }{V: v})
		if err := p.Redraw(ctx); err != nil {
			t.Fatal(err)
		}
	}

	redraw("a")
	redraw("b")
	if s := p.Stats(); s.DirtyRegions != 1 || s.ChangedOps < 1 || s.DirtyOps < 1 {
		t.Fatalf("edit %d/%d/%d", s.DirtyRegions, s.DirtyOps, s.ChangedOps)
	}

	p.SetSize(640, 400)
	if err := p.Redraw(ctx); err != nil {
		t.Fatal(err)
	}
	if s := p.Stats(); s.DirtyRegions != 1 || s.ChangedOps <= 0 {
		t.Fatalf("resize %d/%d", s.DirtyRegions, s.ChangedOps)
	}
}

func TestPerfStageAccounting(t *testing.T) {
	p := newCachePage(t, `<p>stage accounting</p>`)
	p.SetPerf(true)
	if err := p.Redraw(t.Context()); err != nil {
		t.Fatal(err)
	}
	s := p.Stats()
	sum := s.LastTemplate + s.StyleTime + s.LayoutTime + s.PaintTime + s.OtherTime
	if sum != s.LastRedraw {
		t.Fatalf("stage sum %s != redraw %s", sum, s.LastRedraw)
	}
	if s.StyleTime <= 0 || s.LayoutTime <= 0 {
		t.Fatal("style and placement must be measured independently")
	}

	// An identical frame skips the style and placement stages, so they read
	// zero while the stage sum still matches the redraw.
	if err := p.Redraw(t.Context()); err != nil {
		t.Fatal(err)
	}
	s = p.Stats()
	sum = s.LastTemplate + s.StyleTime + s.LayoutTime + s.PaintTime + s.OtherTime
	if sum != s.LastRedraw {
		t.Fatalf("stage sum %s != redraw %s", sum, s.LastRedraw)
	}
	if s.StyleTime != 0 || s.LayoutTime != 0 {
		t.Fatalf("skipped style %s layout %s, want both 0", s.StyleTime, s.LayoutTime)
	}
}
