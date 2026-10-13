package frame

import (
	"testing"
	"time"
)

// BenchmarkAnimCardPatched patches the retained card ops each frame and
// marks the boxes they moved between.
func BenchmarkAnimCardPatched(b *testing.B) {
	screen, ops := benchCard(b)
	span := float64(benchFrames-1) * 0.75

	b.ReportAllocs()

	var a *Anim

	for i := 0; b.Loop(); i++ {
		f := i % benchFrames
		if f == 0 {
			a = TranslateTo(screen.Display(), ops, span, 0, benchFrames*benchStep, Linear)
		}

		a.Tick(screen, t0.Add(time.Duration(f)*benchStep))
	}
}
