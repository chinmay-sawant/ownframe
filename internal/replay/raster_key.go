package replay

import "image/color"

// rasterKey names one cached vector rasterization. It holds everything that
// shapes the pixels: the variant, the size and corner radii, the stroke width
// and mask, the color, and the sub-pixel offset of the op's origin. The
// integer part of the position is not in the key, because the raster bakes at
// the fractional offset and blits at the integer one, so equal shapes at
// different integer positions share one raster.
type rasterKey struct {
	variant uint8
	mask    uint8
	col     color.RGBA
	w, h    float32
	width   float32
	r, ry   [4]float32
	fx, fy  float32
}

// Raster variants.
const (
	rasterFill uint8 = iota + 1
	rasterStroke
)
