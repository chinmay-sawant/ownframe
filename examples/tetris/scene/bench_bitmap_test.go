package scene

import (
	"context"
	"testing"
	"time"
)

// BenchmarkTickBitmap measures one tick against the bitmap fallback page.
func BenchmarkTickBitmap(b *testing.B) {
	s, clock, step := benchScene(b, true)
	if s.page.Testing().Display() != nil {
		b.Skip("the fallback page kept a display list")
	}

	m, ok := s.model.(*fakeModel)
	if !ok {
		b.Fatal("scene model is not a fakeModel")
	}
	ctx := context.Background()

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		m.f.Score = i
		step.advance = 1
		clock.add(time.Second / 60)

		if err := s.Tick(ctx); err != nil {
			b.Fatal(err)
		}
	}
}
