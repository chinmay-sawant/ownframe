package replay

import "image/color"

// textRasterKey names one cached rendered run. It holds everything that
// shapes the pixels: the face identity, the content after text-transform, the
// effective color with its opacity, the fake-bold double strike, the letter
// spacing in pixels, and the sub-pixel phase of the run's draw origin, on a
// 1/16 pixel grid. The integer part of the origin is not in the key, because
// the raster bakes against the run's sub-pixel phase and blits at an integer
// origin, so equal runs at different integer positions share a raster. The
// cache entry carries the baked image's offset from that integer origin, so a
// hit never reshapes the run to find out where the ink lies.
type textRasterKey struct {
	font     any
	content  string
	language string
	size     float64
	spacing  float64
	col      color.RGBA
	fakeBold bool
	fx, fy   float64
}
