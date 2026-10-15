package replay

import (
	"testing"

	"github.com/hajimehoshi/ebiten/v2"

	"github.com/chinmay-sawant/blinkless/layout"
)

// TestZZSleepyEyeRenders replays the exact ticked sleepy eye op and counts
// lit pixels. A mixed-corner fill must paint like any other fill.
func TestZZSleepyEyeRenders(t *testing.T) {
	mk := func(r [4]float64) *layout.DisplayOp {
		return &layout.DisplayOp{Kind: layout.DisplayOpFillRect,
			X: 244.2, Y: 215.5, W: 103.4, H: 53.4,
			R: 0.95, G: 0.96, B: 1, Alpha: 1, Radius: 54,
			RadiusTopLeft: r[0], RadiusTopRight: r[1],
			RadiusBottomRight: r[2], RadiusBottomLeft: r[3]}
	}
	lit := func(op *layout.DisplayOp) int {
		img := ebiten.NewImage(300, 160)
		defer img.Dispose()
		drawOp(img, op, -200, -180)
		px := make([]byte, 300*160*4)
		img.ReadPixels(px)
		n := 0
		for i := 0; i < len(px); i += 4 {
			if px[i] > 0x80 {
				n++
			}
		}
		return n
	}
	mixed := lit(mk([4]float64{9, 9, 26.16, 26.16}))
	plain := lit(mk([4]float64{}))
	t.Logf("mixed=%d plain=%d", mixed, plain)
	if mixed == 0 {
		t.Fatal("mixed-corner fill painted nothing")
	}
	if plain == 0 {
		t.Fatal("uniform fill painted nothing")
	}
}
