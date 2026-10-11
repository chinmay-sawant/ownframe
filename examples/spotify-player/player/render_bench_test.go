package player

import (
	"context"
	"testing"
)

func BenchmarkScreenSetDataRedraw(b *testing.B) {
	for _, screen := range []string{
		"home", "search", "library", "liked", "browse", "radio", "queue", "profile",
	} {
		b.Run(screen, func(b *testing.B) {
			app, _ := newFakeApp(b)
			app.view.Nav = screen
			app.page.SetPerf(true)
			app.page.SetData(app.view)
			if err := app.Redraw(context.Background()); err != nil {
				b.Fatal(err)
			}
			before := app.page.Stats().Parses
			b.ReportAllocs()
			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				app.page.SetData(app.view)
				if err := app.Redraw(context.Background()); err != nil {
					b.Fatal(err)
				}
			}
			b.StopTimer()
			stats := app.page.Stats()
			parses := stats.Parses - before
			b.ReportMetric(float64(parses)/float64(b.N), "parses/op")
			b.ReportMetric(float64(stats.LayoutTime.Nanoseconds()), "layout-ns/op")
			b.ReportMetric(float64(stats.DisplayListTime.Nanoseconds()), "display-ns/op")
			b.ReportMetric(float64(stats.Boxes), "boxes/page")
			b.ReportMetric(float64(stats.Ops), "ops/page")
		})
	}
}
