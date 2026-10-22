package replay

import (
	"image"
	"math"

	"github.com/chinmay-sawant/blinkless/layout"
	"github.com/hajimehoshi/ebiten/v2"
)

// maxScratchPixels caps one scratch raster, mirroring maxXformPixels.
const maxScratchPixels = 4 << 20

// xformScratch returns the local raster's integer origin and size, or ok
// false when the local box cannot make an image.
func xformScratch(op *layout.DisplayOp) (left, top int, size image.Point, ok bool) {
	l, t, r, b, ok := xformLocalBox(op)
	if !ok {
		return 0, 0, image.Point{}, false
	}

	left = int(math.Floor(l)) - xformPad
	top = int(math.Floor(t)) - xformPad
	right := int(math.Ceil(r)) + xformPad
	bottom := int(math.Ceil(b)) + xformPad
	size = image.Pt(right-left, bottom-top)

	if size.X <= 0 || size.Y <= 0 || oversizedImage(size) {
		return 0, 0, image.Point{}, false
	}

	return left, top, size, true
}

// oversizedImage reports a raster no Ebiten image can hold. A zero limit
// means the game is not running, where no GPU image exists to overflow.
func oversizedImage(size image.Point) bool {
	if size.X <= 0 || size.Y <= 0 {
		return true
	}
	limit := ebiten.MaxImageSize()
	if limit > 0 && (size.X > limit || size.Y > limit) {
		return true
	}
	return int64(size.X)*int64(size.Y) > maxScratchPixels
}
