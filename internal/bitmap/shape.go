package bitmap

import (
	"image"
	"math"

	"github.com/chinmay-sawant/blinkless/layout"
)

func paintShape(img *image.NRGBA, op *layout.DisplayOp) {
	paintCoverage(img, op, shapeBounds(op), func(x, y float64) bool {
		return shapeContains(op, x, y)
	})
}

// paintCoverage paints one operation's ink at four subsamples per pixel,
// scaling the colour's alpha by how many subsamples land inside. hit reports
// whether one point in the operation's local space is ink; the caller bounds
// the loop because a rounded rectangle and a shaped path bound themselves
// differently.
func paintCoverage(img *image.NRGBA, op *layout.DisplayOp, bounds image.Rectangle,
	hit func(x, y float64) bool) {
	bounds = bounds.Intersect(img.Bounds())
	c := opColor(op)

	for y := bounds.Min.Y; y < bounds.Max.Y; y++ {
		for x := bounds.Min.X; x < bounds.Max.X; x++ {
			count := 0
			for _, oy := range [2]float64{.25, .75} {
				for _, ox := range [2]float64{.25, .75} {
					px, py, ok := untransform(op, (float64(x)+ox)/pxPerPt, (float64(y)+oy)/pxPerPt)
					if ok && hit(px, py) {
						count++
					}
				}
			}
			if count == 0 {
				continue
			}
			src := c
			src.A = uint8(float64(src.A) * float64(count) / shapeSamples)
			img.SetNRGBA(x, y, over(img.NRGBAAt(x, y), src))
		}
	}
}

func shapeBounds(op *layout.DisplayOp) image.Rectangle {
	x, y, w, h := op.X, op.Y, op.W, op.H
	if op.Kind == layout.DisplayOpLine {
		x, y, w, h, _ = op.PaintLineGeometry()
	}
	pad := max(op.Width, 1)
	x0, y0, x1, y1 := min(x, x+w)-pad, min(y, y+h)-pad, max(x, x+w)+pad, max(y, y+h)+pad
	loX, loY, hiX, hiY := math.Inf(1), math.Inf(1), math.Inf(-1), math.Inf(-1)
	for _, p := range [][2]float64{{x0, y0}, {x1, y0}, {x1, y1}, {x0, y1}} {
		xx, yy := transform(op, p[0], p[1])
		loX, loY, hiX, hiY = min(loX, xx), min(loY, yy), max(hiX, xx), max(hiY, yy)
	}
	return image.Rect(int(math.Floor(loX*pxPerPt)), int(math.Floor(loY*pxPerPt)),
		int(math.Ceil(hiX*pxPerPt)), int(math.Ceil(hiY*pxPerPt)))
}
