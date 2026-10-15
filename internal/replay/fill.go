package replay

import (
	"math"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/vector"

	"github.com/chinmay-sawant/blinkless/layout"

	"github.com/chinmay-sawant/ownframe/internal/masking"
)

// fillRect draws one filled rectangle. Corners are circular; the caller has
// already rejected elliptical corners. Plain fills snap to whole pixels, the
// same geometry the engine's raster painter uses.
func fillRect(dst *ebiten.Image, op *layout.DisplayOp, dx, dy float64) {
	radii, ok := masking.FillRadii(op)
	if !ok {
		return
	}

	if radii == ([4]float64{}) {
		left, top, right, bottom := snapped(op, dx, dy)
		vector.FillRect(dst, left, top, right-left, bottom-top, rgba(op), false)

		return
	}

	// A repeated rounded shape comes from the raster cache; a shape too large
	// to cache falls through to the direct path.
	if cachedRoundedFill(dst, op, radii, dx, dy) {
		return
	}

	fillRounded(dst, op, radii, dx, dy)
}

// snapped converts an op's rectangle to whole canvas pixels, matching the
// engine's ptRectScale rounding.
func snapped(op *layout.DisplayOp, dx, dy float64) (left, top, right, bottom float32) {
	return float32(math.Round(op.X*pxPerPt + dx)), float32(math.Round(op.Y*pxPerPt + dy)),
		float32(math.Round((op.X+op.W)*pxPerPt + dx)), float32(math.Round((op.Y+op.H)*pxPerPt + dy))
}
