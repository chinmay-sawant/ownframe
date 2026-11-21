package main

import (
	"context"
	"math"
	"testing"
	"time"

	"github.com/chinmay-sawant/ownframe"
)

// TestLoglensThemeToggleMidTransitionHover reproduces the regression where
// hovering any control during the 150ms Light->Dark theme transition cancelled
// the other controls' CSS transitions and froze light-theme fills behind dark
// theme text.
func TestLoglensThemeToggleMidTransitionHover(t *testing.T) {
	ctx := context.Background()

	fresh, err := newApp()
	if err != nil {
		t.Fatal(err)
	}
	if err := fresh.page.Redraw(ctx); err != nil {
		t.Fatal(err)
	}

	a, err := newApp()
	if err != nil {
		t.Fatal(err)
	}
	if err := a.page.Redraw(ctx); err != nil {
		t.Fatal(err)
	}

	findBox := func(p *ownframe.Page, target string) ownframe.Box {
		t.Helper()
		for _, b := range p.Boxes() {
			if b.Text == target {
				return b
			}
		}
		t.Fatalf("missing box %q", target)
		return ownframe.Box{}
	}

	click := func(p *ownframe.Page, target string) {
		t.Helper()
		b := findBox(p, target)
		p.Probe().BeginRedrawBatch()
		_ = p.Edit().Hover(ctx, b.X+b.W/2, b.Y+b.H/2)
		_ = p.Edit().Press(ctx, b.X+b.W/2, b.Y+b.H/2)
		_ = p.Click(ctx, b.X+b.W/2, b.Y+b.H/2)
		_ = p.Edit().Release(ctx)
		if err := p.Probe().EndRedrawBatch(ctx); err != nil {
			t.Fatal(err)
		}
		_ = p.Testing().Tick(ctx)
	}

	hover := func(p *ownframe.Page, target string) {
		t.Helper()
		b := findBox(p, target)
		p.Probe().BeginRedrawBatch()
		if err := p.Edit().Hover(ctx, b.X+b.W/2, b.Y+b.H/2); err != nil {
			t.Fatal(err)
		}
		if err := p.Probe().EndRedrawBatch(ctx); err != nil {
			t.Fatal(err)
		}
		_ = p.Testing().Tick(ctx)
	}

	// Switch Dark -> Light (let settle), then Light -> Dark, and hover "Traces"
	// 20ms into the 150ms theme transition.
	click(a.page, "Light")
	time.Sleep(220 * time.Millisecond)
	_ = a.page.Testing().Tick(ctx)

	click(a.page, "Dark")
	time.Sleep(20 * time.Millisecond)
	_ = a.page.Testing().Tick(ctx)
	hover(a.page, "Traces")

	// Let every transition finish.
	time.Sleep(250 * time.Millisecond)
	_ = a.page.Testing().Tick(ctx)

	// Match hover state on fresh dark baseline.
	hover(fresh.page, "Traces")
	time.Sleep(250 * time.Millisecond)
	_ = fresh.page.Testing().Tick(ctx)

	dFresh := fresh.page.Probe().Display()
	dGot := a.page.Probe().Display()
	if len(dFresh.Ops) != len(dGot.Ops) {
		t.Fatalf("op count mismatch after Dark transition settled: fresh=%d got=%d", len(dFresh.Ops), len(dGot.Ops))
	}
	colorDiffs := 0
	for i := range dFresh.Ops {
		f, g := &dFresh.Ops[i], &dGot.Ops[i]
		if f.Kind != g.Kind {
			t.Fatalf("op[%d] kind mismatch: fresh=%v got=%v", i, f.Kind, g.Kind)
		}
		if math.Abs(f.R-g.R) > 0.02 || math.Abs(f.G-g.G) > 0.02 ||
			math.Abs(f.B-g.B) > 0.02 || math.Abs(f.Alpha-g.Alpha) > 0.02 {
			if colorDiffs < 10 {
				t.Errorf("op[%d] (%v %q @ %.0f,%.0f) color mismatch: fresh=(%.2f,%.2f,%.2f,%.2f) got=(%.2f,%.2f,%.2f,%.2f)",
					i, f.Kind, f.Text, f.X, f.Y, f.R, f.G, f.B, f.Alpha, g.R, g.G, g.B, g.Alpha)
			}
			colorDiffs++
		}
	}
	if colorDiffs > 0 {
		t.Fatalf("total color mismatches after mid-transition hover = %d", colorDiffs)
	}

	// Moving hover from Platform to Live (which runs PatchElements on the live
	// tooltip) and then to Traces must leave Platform cleanly un-hovered.
	hover(a.page, "Platform")
	hover(a.page, "Live")
	hover(a.page, "Traces")
	time.Sleep(250 * time.Millisecond)
	_ = a.page.Testing().Tick(ctx)
	if got, want := len(a.page.Probe().Display().Ops), len(dFresh.Ops); got != want {
		t.Fatalf("op count after Platform->Live->Traces = %d, want %d (stale hover fill left behind)", got, want)
	}
}
