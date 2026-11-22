package bitmap

import (
	"image"
	"image/color"
	"math"

	"golang.org/x/image/font"
	"golang.org/x/image/math/fixed"

	"github.com/chinmay-sawant/blinkless/layout"
)

// strokeRingPoints is how many offsets one ring of the outline uses.
const strokeRingPoints = 12

// strokeHalfWidthDivisor turns a stroke width into the radius of its centre ring.
// wideStrokeRadius is the radius in pixels above which a smaller inner ring is
// drawn too, and innerRingDivisor sets that inner ring to half the radius.
const (
	strokeHalfWidthDivisor = 2
	wideStrokeRadius       = 1.5
	innerRingDivisor       = 2
)

// paintTextStroke paints the -webkit-text-stroke outline under the fill by
// drawing the run in the stroke color on a circle of half the stroke width,
// and on a smaller ring for wide strokes. The caller draws the fill after, so
// only the outer half of the stroke shows. A real stroke is centered on the
// glyph edge and paints over the fill.
func paintTextStroke(dst *image.NRGBA, op *layout.DisplayOp, face font.Face, content string, x, y float64) {
	width, rgb, ok := op.TextStroke()
	if !ok {
		return
	}

	radius := width * pxPerPt / strokeHalfWidthDivisor
	rings := []float64{radius}

	if radius > wideStrokeRadius {
		rings = append(rings, radius/innerRingDivisor)
	}

	for _, ring := range rings {
		for step := range strokeRingPoints {
			angle := 2 * math.Pi * float64(step) / strokeRingPoints
			draw := font.Drawer{
				Dst: dst,
				Src: image.NewUniform(color.NRGBA{
					R: channel(rgb[0]), G: channel(rgb[1]), B: channel(rgb[2]), A: channel(op.Opacity()),
				}),
				Face: face,
				Dot: fixed.Point26_6{
					X: fixed.Int26_6(math.Round((x*pxPerPt + ring*math.Cos(angle)) * fixedPointScale)),
					Y: fixed.Int26_6(math.Round((y*pxPerPt + ring*math.Sin(angle)) * fixedPointScale)),
				},
			}
			drawString(&draw, content, op.LetterSpacing*pxPerPt, op.TextAutospaceGap()*pxPerPt)
		}
	}
}
