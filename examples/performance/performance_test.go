package main

import (
	"context"
	"testing"

	"github.com/chinmay-sawant/ownframe"
)

func TestAllScenariosRender(t *testing.T) {
	ctx := context.Background()
	scenarios := AllScenarios()

	if len(scenarios) != 10 {
		t.Fatalf("expected 10 scenarios, got %d", len(scenarios))
	}

	for _, s := range scenarios {
		t.Run(s.ID, func(t *testing.T) {
			page, err := ownframe.New(ownframe.Config{
				Title:  s.Name,
				HTML:   s.HTML,
				Width:  s.Width,
				Height: s.Height,
				Perf:   true,
			})
			if err != nil {
				t.Fatalf("[%s] page creation failed: %v", s.ID, err)
			}

			page.SetData(s.Data)

			if err := page.Redraw(ctx); err != nil {
				t.Fatalf("[%s] initial Redraw failed: %v", s.ID, err)
			}

			stats := page.Probe().Stats()
			if stats.Ops <= 0 {
				t.Errorf("[%s] expected > 0 ops, got %d", s.ID, stats.Ops)
			}

			boxesCount := 0
			if d := page.Probe().Display(); d != nil {
				boxesCount = len(d.Boxes)
			} else {
				boxesCount = len(page.Boxes())
			}
			if boxesCount <= 0 {
				t.Errorf("[%s] expected > 0 boxes, got %d", s.ID, boxesCount)
			}

			// Test warm redraw
			if err := page.Redraw(ctx); err != nil {
				t.Errorf("[%s] warm Redraw failed: %v", s.ID, err)
			}

			// Test resize
			page.SetSize(s.Width+2, s.Height+2)
			if err := page.Redraw(ctx); err != nil {
				t.Errorf("[%s] resize Redraw failed: %v", s.ID, err)
			}
		})
	}
}

// -----------------------------------------------------------------------------
// Benchmark Suite for 10 Scenarios (Warm Redraw)
// -----------------------------------------------------------------------------

func runWarmBenchmark(b *testing.B, s Scenario) {
	ctx := context.Background()
	page, err := ownframe.New(ownframe.Config{
		Title:  s.Name,
		HTML:   s.HTML,
		Width:  s.Width,
		Height: s.Height,
		Perf:   false,
	})
	if err != nil {
		b.Fatal(err)
	}

	page.SetData(s.Data)
	if err := page.Redraw(ctx); err != nil {
		b.Fatal(err)
	}

	b.ResetTimer()
	b.ReportAllocs()

	for b.Loop() {
		if err := page.Redraw(ctx); err != nil {
			b.Fatal(err)
		}
	}
}

func runColdBenchmark(b *testing.B, s Scenario) {
	ctx := context.Background()
	b.ReportAllocs()

	for b.Loop() {
		page, err := ownframe.New(ownframe.Config{
			Title:  s.Name,
			HTML:   s.HTML,
			Width:  s.Width,
			Height: s.Height,
			Perf:   false,
		})
		if err != nil {
			b.Fatal(err)
		}

		page.SetData(s.Data)
		if err := page.Redraw(ctx); err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkScenario01_SimpleRows50_Warm(b *testing.B) {
	runWarmBenchmark(b, makeScenario1SimpleRows50())
}
func BenchmarkScenario01_SimpleRows50_Cold(b *testing.B) {
	runColdBenchmark(b, makeScenario1SimpleRows50())
}

func BenchmarkScenario02_ZebraTable200_Warm(b *testing.B) {
	runWarmBenchmark(b, makeScenario2ZebraTable200())
}
func BenchmarkScenario02_ZebraTable200_Cold(b *testing.B) {
	runColdBenchmark(b, makeScenario2ZebraTable200())
}

func BenchmarkScenario03_ProductGrid200_Warm(b *testing.B) {
	runWarmBenchmark(b, makeScenario3ProductGrid200())
}
func BenchmarkScenario03_ProductGrid200_Cold(b *testing.B) {
	runColdBenchmark(b, makeScenario3ProductGrid200())
}

func BenchmarkScenario04_SocialTimeline200_Warm(b *testing.B) {
	runWarmBenchmark(b, makeScenario4SocialTimeline200())
}
func BenchmarkScenario04_SocialTimeline200_Cold(b *testing.B) {
	runColdBenchmark(b, makeScenario4SocialTimeline200())
}

func BenchmarkScenario05_OrderBook500_Warm(b *testing.B) {
	runWarmBenchmark(b, makeScenario5OrderBook500())
}
func BenchmarkScenario05_OrderBook500_Cold(b *testing.B) {
	runColdBenchmark(b, makeScenario5OrderBook500())
}

func BenchmarkScenario06_EnterpriseGrid1000_Warm(b *testing.B) {
	runWarmBenchmark(b, makeScenario6EnterpriseGrid1000())
}
func BenchmarkScenario06_EnterpriseGrid1000_Cold(b *testing.B) {
	runColdBenchmark(b, makeScenario6EnterpriseGrid1000())
}

func BenchmarkScenario07_AnalyticsDashboard_Warm(b *testing.B) {
	runWarmBenchmark(b, makeScenario7AnalyticsDashboard())
}
func BenchmarkScenario07_AnalyticsDashboard_Cold(b *testing.B) {
	runColdBenchmark(b, makeScenario7AnalyticsDashboard())
}

func BenchmarkScenario08_MultiPageDoc200_Warm(b *testing.B) {
	runWarmBenchmark(b, makeScenario8MultiPageDoc200())
}
func BenchmarkScenario08_MultiPageDoc200_Cold(b *testing.B) {
	runColdBenchmark(b, makeScenario8MultiPageDoc200())
}

func BenchmarkScenario09_CSSKitchenSink150_Warm(b *testing.B) {
	runWarmBenchmark(b, makeScenario9CSSKitchenSink150())
}
func BenchmarkScenario09_CSSKitchenSink150_Cold(b *testing.B) {
	runColdBenchmark(b, makeScenario9CSSKitchenSink150())
}

func BenchmarkScenario10_ExtremeApp1000_Warm(b *testing.B) {
	runWarmBenchmark(b, makeScenario10ExtremeApp1000())
}
func BenchmarkScenario10_ExtremeApp1000_Cold(b *testing.B) {
	runColdBenchmark(b, makeScenario10ExtremeApp1000())
}
