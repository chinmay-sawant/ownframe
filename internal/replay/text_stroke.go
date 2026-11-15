package replay

import (
	"math"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/text/v2"

	"github.com/chinmay-sawant/blinkless/layout"
)

// strokeRingPoints is how many offsets one ring of the outline uses.
const strokeRingPoints = 12

// drawTextStroke paints the -webkit-text-stroke outline under the fill. The
// engine carries the stroke width and color on the op (Op.TextStroke). Ebiten
// text cannot stroke a glyph path, so the run is drawn in the stroke color at
// points around a circle of half the stroke width, and again on a smaller
// ring for wide strokes. The fill is drawn after this, so only the half of
// the stroke outside the glyph shows. A real stroke is centered on the glyph
// edge and paints over the fill, so this outline is about half as heavy.
func (renderer *Renderer) drawTextStroke(
	dst *ebiten.Image, face *text.GoTextFace, content string, x, y float64, op *layout.DisplayOp,
) {
	width, color, ok := op.TextStroke()
	if !ok {
		return
	}

	stroked := *op
	stroked.R, stroked.G, stroked.B = color[0], color[1], color[2]

	radius := width * pxPerPt / 2
	rings := []float64{radius}

	if radius > 1.5 {
		rings = append(rings, radius/2)
	}

	for _, ring := range rings {
		for step := range strokeRingPoints {
			angle := 2 * math.Pi * float64(step) / strokeRingPoints
			renderer.drawTextRun(dst, face, content, x+ring*math.Cos(angle), y+ring*math.Sin(angle), &stroked)
		}
	}
}
