package replay

import (
	"image"
	"testing"

	"github.com/chinmay-sawant/blinkless/layout"
	"github.com/hajimehoshi/ebiten/v2"
)

// BenchmarkChunkCardDamage damages one card of a 20-card grid and draws it
// into an image, direct and chunked, reporting the draws and time per call.
func BenchmarkChunkCardDamage(b *testing.B) {
	d := cardBenchDisplay()
	rect := image.Rect(benchGap-2, benchGap-2,
		benchGap+benchCardW+2, benchGap+benchCardH+2)
	dst := ebiten.NewImage(d.Width, d.Height)
	defer dst.Deallocate()

	replayCounts.on.Store(true)
	defer replayCounts.on.Store(false)

	_, runs, members := planStats(chunkPlanFor(d))
	b.Logf("ops=%d runs=%d run-ops=%d", len(d.Ops), runs, members)

	b.Run("direct", func(b *testing.B) {
		b.ReportAllocs()
		replayCounts.ops.Store(0)
		replayCounts.blits.Store(0)
		b.ResetTimer()

		for i := 0; i < b.N; i++ {
			drawDirectRect(dst, d, rect)
		}

		b.StopTimer()
		b.ReportMetric(float64(replayCounts.ops.Load())/float64(b.N), "ops/draw")
		b.ReportMetric(float64(replayCounts.blits.Load())/float64(b.N), "blits/draw")
	})

	b.Run("chunked", func(b *testing.B) {
		b.ReportAllocs()
		replayCounts.ops.Store(0)
		replayCounts.blits.Store(0)
		DrawRect(dst, d, rect, 0, 0)
		bakeOps := replayCounts.ops.Load()
		bakeBlits := replayCounts.blits.Load()
		b.ResetTimer()

		for i := 0; i < b.N; i++ {
			DrawRect(dst, d, rect, 0, 0)
		}

		b.StopTimer()
		b.ReportMetric(float64(replayCounts.ops.Load()-bakeOps)/float64(b.N), "ops/draw")
		b.ReportMetric(float64(replayCounts.blits.Load()-bakeBlits)/float64(b.N), "blits/draw")
		b.Logf("bake ops=%d", bakeOps)
	})
}

// drawDirectRect is the pre-chunk damage path: filter the order and draw
// every op that touches the rect.
func drawDirectRect(dst *ebiten.Image, d *layout.Display, rect image.Rectangle) {
	ppt := d.PointsPerPixel

	for _, i := range d.Order {
		op := &d.Ops[i]
		if op.Fixed || !opTouches(op, ppt, rect) {
			continue
		}

		countOp()
		drawOp(dst, op, 0, 0)
	}
}
