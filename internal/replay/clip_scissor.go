package replay

import (
	"image"
	"math"

	"github.com/hajimehoshi/ebiten/v2"

	"github.com/chinmay-sawant/blinkless/layout"
)

// scissorOverflow returns dst clipped to the op's recorded overflow box.
// Geometry-baked clipping shrinks the op bounds, which is exact for static
// paint but moves with a sampled matrix; a host painting through a transform
// must intersect the paint with the ancestor clip instead. It reports false
// when the clip erases the paint so the caller skips the draw. Ops without a
// recorded clip keep dst unchanged.
func scissorOverflow(dst *ebiten.Image, op *layout.DisplayOp, dx, dy float64) (*ebiten.Image, bool) {
	cx, cy, cw, ch, ok := op.OverflowClip()
	if !ok || dst == nil {
		return dst, true
	}

	r := image.Rect(
		int(math.Floor(cx*pxPerPt+dx)),
		int(math.Floor(cy*pxPerPt+dy)),
		int(math.Ceil((cx+cw)*pxPerPt+dx)),
		int(math.Ceil((cy+ch)*pxPerPt+dy)),
	).Intersect(dst.Bounds())

	if r.Empty() {
		return dst, false
	}

	if r.Eq(dst.Bounds()) {
		return dst, true
	}

	sub, ok := dst.SubImage(r).(*ebiten.Image)
	if !ok {
		return dst, true
	}

	return sub, true
}
