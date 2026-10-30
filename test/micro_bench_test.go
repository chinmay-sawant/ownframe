package test_test

import (
	"context"
	"testing"
	"time"

	"github.com/chinmay-sawant/ownframe"
	"github.com/chinmay-sawant/ownframe/pagetest"
)

// ----------------------------------------------------------------------------
// Category 2: Optimized Microbenchmark Suite
// ----------------------------------------------------------------------------

func TestMicro_TickLatency(t *testing.T) {
	ctx := context.Background()
	p, err := ownframe.NewWithOptions(
		ownframe.Config{HTML: cardHTML, Width: 400, Height: 300},
		ownframe.WithPerf(true),
	)
	if err != nil {
		t.Fatal(err)
	}
	if err := p.Redraw(ctx); err != nil {
		t.Fatal(err)
	}
	p.Probe().TakeDirty()

	tickCount := 0
	p.SetTick(func(context.Context) error {
		tickCount++
		return nil
	})

	const iterations = 100
	t0 := time.Now()
	for i := 0; i < iterations; i++ {
		if err := p.Testing().Tick(ctx); err != nil {
			t.Fatalf("Tick failed: %v", err)
		}
	}
	totalElapsed := time.Since(t0)
	avgPerTick := totalElapsed / iterations

	t.Logf("=== 60 FPS Game Loop Idle Tick Cost ===")
	t.Logf("100 ticks completed in %v (Average: %v per tick frame)", totalElapsed, avgPerTick)

	if avgPerTick > 1*time.Millisecond {
		t.Errorf("Average idle tick %v is too slow (> 1ms)", avgPerTick)
	}
}

func TestMicro_ClickHandlingLatency(t *testing.T) {
	ctx := context.Background()
	p, err := ownframe.NewWithOptions(
		ownframe.Config{HTML: cardHTML, Width: 400, Height: 300},
		ownframe.WithPerf(true),
	)
	if err != nil {
		t.Fatal(err)
	}

	clicked := false
	p.Handle(ownframe.Handlers{
		Click: func(_ context.Context, b ownframe.Box) error {
			if b.Action == "click-test" {
				clicked = true
				p.Probe().Invalidate("btn")
			}
			return nil
		},
	})

	if err := p.Redraw(ctx); err != nil {
		t.Fatal(err)
	}
	p.Probe().TakeDirty()

	t0 := time.Now()
	if err := pagetest.ClickAction(ctx, p, "click-test"); err != nil {
		t.Fatalf("ClickAction failed: %v", err)
	}
	clickDuration := time.Since(t0)

	if !clicked {
		t.Fatal("click handler was not triggered")
	}

	dirtyRect, hasDirty := p.Probe().TakeDirty()
	if !hasDirty || dirtyRect.Empty() {
		t.Fatal("expected dirty bounding box after invalidation")
	}

	t.Logf("=== Interactive Click Dispatch Latency ===")
	t.Logf("Click hit-test, handler invocation & dirty marking took: %v", clickDuration)
	t.Logf("Dirty box marked for GPU partial repaint: %v", dirtyRect)
}

func BenchmarkFrame_Tick_Idle(b *testing.B) {
	ctx := context.Background()
	p, err := ownframe.New(ownframe.Config{HTML: dashboardLayoutHTML, Width: 800, Height: 600})
	if err != nil {
		b.Fatal(err)
	}
	_ = p.Redraw(ctx)

	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = p.Testing().Tick(ctx)
	}
}

func BenchmarkFrame_WarmRedraw_Card(b *testing.B) {
	ctx := context.Background()
	p, err := ownframe.New(ownframe.Config{HTML: cardHTML, Width: 400, Height: 300})
	if err != nil {
		b.Fatal(err)
	}
	_ = p.Redraw(ctx)

	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = p.RedrawIfChanged(ctx)
	}
}

func BenchmarkFrame_WarmRedraw_Dashboard(b *testing.B) {
	ctx := context.Background()
	p, err := ownframe.New(ownframe.Config{HTML: dashboardLayoutHTML, Width: 800, Height: 600})
	if err != nil {
		b.Fatal(err)
	}
	_ = p.Redraw(ctx)

	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = p.RedrawIfChanged(ctx)
	}
}

func BenchmarkFrame_Dirty_InvalidateAndRedraw(b *testing.B) {
	ctx := context.Background()
	p, err := ownframe.NewWithOptions(
		ownframe.Config{HTML: cardHTML, Width: 400, Height: 300},
		ownframe.WithPerf(true),
	)
	if err != nil {
		b.Fatal(err)
	}
	_ = p.Redraw(ctx)
	p.Probe().TakeDirty()

	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		p.Probe().Invalidate("btn")
		p.Probe().TakeDirty()
	}
}

func BenchmarkFrame_Click_Dispatch(b *testing.B) {
	ctx := context.Background()
	p, err := ownframe.New(ownframe.Config{HTML: cardHTML, Width: 400, Height: 300})
	if err != nil {
		b.Fatal(err)
	}
	p.Handle(ownframe.Handlers{
		Click: func(_ context.Context, b ownframe.Box) error {
			return nil
		},
	})
	_ = p.Redraw(ctx)

	box, ok := pagetest.BoxByAction(p, "click-test")
	if !ok {
		b.Fatal("missing action box")
	}
	cx := box.X + box.W/2
	cy := box.Y + box.H/2

	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = p.Click(ctx, cx, cy)
	}
}
