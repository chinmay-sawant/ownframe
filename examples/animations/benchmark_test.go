package main

import (
	"context"
	"testing"
)

// BenchmarkAnimationTick samples the whole gallery, including offscreen 3D
// cards, without the window's viewport culling, replay, or DevTools overhead.
func BenchmarkAnimationTick(b *testing.B) {
	app, err := newApp(false, false)
	if err != nil {
		b.Fatal(err)
	}
	p := app.Page()
	ctx := context.Background()
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		if err := p.Testing().Tick(ctx); err != nil {
			b.Fatal(err)
		}
		p.Probe().TakeDirtyRects()
	}
}
