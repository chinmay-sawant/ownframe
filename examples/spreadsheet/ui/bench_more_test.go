package ui

import (
	"context"
	"testing"
)

// BenchmarkRedrawWindowReplace crosses an overscan edge every step.
func BenchmarkRedrawWindowReplace(b *testing.B) {
	app := benchApp(b)
	ctx := context.Background()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		app.page.SetScrollOffset(0, (i%40)*RowH*8)
		app.page.Testing().StepScrollWindow()
		if err := app.page.Redraw(ctx); err != nil {
			b.Fatal(err)
		}
	}
}

// BenchmarkTickApply is one tick that applies a full-window fetch result
// and redraws: the active-tick cost the UI pays when data lands.
func BenchmarkTickApply(b *testing.B) {
	app := benchApp(b)
	ctx := context.Background()
	a := app.win.Area()
	full := make([]Cell, a.Count())
	for i := range full {
		full[i] = Cell{Raw: "v", Display: "v"}
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		app.work.out <- result{kind: jobFetch, gen: app.fetchGen, sheet: "s1", area: a, cells: full}
		if err := app.tick(ctx); err != nil {
			b.Fatal(err)
		}
	}
}

// BenchmarkRedrawSameSource relayouts the cached document with no change.
func BenchmarkRedrawSameSource(b *testing.B) {
	app := benchApp(b)
	ctx := context.Background()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		if err := app.page.Redraw(ctx); err != nil {
			b.Fatal(err)
		}
	}
}
