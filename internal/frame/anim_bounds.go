package frame

import (
	"image"
	"math"

	"github.com/chinmay-sawant/ownframe"
)

// OpBox returns op's painted box in CSS pixels: the box to mark when the op
// moves, fades, or changes shape. It honors a baked transform and the text
// baseline, matching the page's own dirty boxes.
func OpBox(d *ownframe.Display, op *ownframe.DisplayOp) image.Rectangle {
	x0, y0, x1, y1 := localBox(op)

	m := op.Transform()
	if op.XformSet && !m.IsIdentity() {
		x0, y0, x1, y1 = mapBox(m.A, m.B, m.C, m.D, m.E, m.F, x0, y0, x1, y1)
	}

	return pxBox(d, x0, y0, x1, y1)
}

// pxBox converts a point box to CSS pixels with a pixel of slack.
func pxBox(d *ownframe.Display, x0, y0, x1, y1 float64) image.Rectangle {
	s := 1.0
	if d != nil && d.PixelPerPoint > 0 {
		s = d.PixelPerPoint
	}

	return image.Rect(int(x0*s)-1, int(y0*s)-1, int(x1*s)+1, int(y1*s)+1)
}

// mapBox is the bounding box of the box's four corners under a, b, c, d, e, f.
func mapBox(a, b, c, d, e, f, x0, y0, x1, y1 float64) (float64, float64, float64, float64) {
	loX, loY := math.Inf(1), math.Inf(1)
	hiX, hiY := math.Inf(-1), math.Inf(-1)

	for _, p := range [4][2]float64{{x0, y0}, {x1, y0}, {x1, y1}, {x0, y1}} {
		x, y := a*p[0]+c*p[1]+e, b*p[0]+d*p[1]+f
		loX, loY = math.Min(loX, x), math.Min(loY, y)
		hiX, hiY = math.Max(hiX, x), math.Max(hiY, y)
	}

	return loX, loY, hiX, hiY
}
