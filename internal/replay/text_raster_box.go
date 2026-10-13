package replay

import (
	"image"
	"math"
)

// textRasterMaxSide, textRasterMaxPixels, and textRasterPad bound one baked
// run. Pad is the integer slack per side around the glyph ink: it covers the
// sub-pixel placement of the glyph images and the rounding the face does when
// it quantizes the run origin to its sub-pixel grid. Past the caps the run
// draws directly.
const (
	textRasterMaxSide   = 512
	textRasterMaxPixels = 1 << 20
	textRasterPad       = 4
)

// textPhaseGrid is the sub-pixel grid the cache keys runs on. Snapping the
// offset to it absorbs the float64 noise that an integer scroll shift leaves
// in the origin, so a run scrolled by whole pixels keeps its key. The bake
// still uses the exact first origin, so a hit is a phase match to within
// 1/32 of a pixel.
const textPhaseGrid = 16

// textPhase snaps a sub-pixel offset to the keying grid.
func textPhase(frac float64) float64 {
	return math.Round(frac*textPhaseGrid) / textPhaseGrid
}

// textRasterLayout pads the run's ink box and returns the image size and the
// image's offset from the run origin. ok is false when the run is past the
// side or pixel caps and must draw directly.
func textRasterLayout(ink image.Rectangle) (size image.Point, dx, dy int, ok bool) {
	ink = ink.Inset(-textRasterPad)
	if ink.Dx() > textRasterMaxSide || ink.Dy() > textRasterMaxSide ||
		ink.Dx()*ink.Dy() > textRasterMaxPixels {
		return image.Point{}, 0, 0, false
	}

	return image.Pt(ink.Dx(), ink.Dy()), ink.Min.X, ink.Min.Y, true
}
