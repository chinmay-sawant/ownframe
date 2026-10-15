package textrun

import (
	"math"

	"github.com/hajimehoshi/ebiten/v2/text/v2"

	"github.com/chinmay-sawant/blinkless/layout"
)

// runWidth measures a run in CSS pixels.
func runWidth(op *layout.DisplayOp, pxPerPt float64) float64 {
	face := faceFor(op, pxPerPt)
	if face == nil {
		return op.W * pxPerPt
	}

	w, _ := text.Measure(drawn(op), face, 0)

	return w
}

// offsetIn returns the rune offset nearest dxPx from the run's start.
// dxPx is in CSS pixels.
func offsetIn(op *layout.DisplayOp, dxPx, pxPerPt float64) int {
	runes := []rune(drawn(op))
	if len(runes) == 0 {
		return 0
	}

	face := faceFor(op, pxPerPt)
	if face == nil {
		return len(runes)
	}

	spacing := op.LetterSpacing * pxPerPt
	best, bestDist := 0, math.Abs(dxPx)
	width := 0.0

	// Widths grow monotonically when spacing is not negative, so the scan
	// stops once past the target by more than the best gap. Skipped
	// iterations cannot beat best.
	monotonic := spacing >= 0
	for i := 1; i <= len(runes); i++ {
		if spacing != 0 {
			glyph, _ := text.Measure(string(runes[i-1]), face, 0)
			width += glyph + spacing
		} else {
			width, _ = text.Measure(string(runes[:i]), face, 0)
		}

		if d := math.Abs(width - dxPx); d < bestDist {
			best, bestDist = i, d
		} else if monotonic && width > dxPx+bestDist {
			break
		}
	}

	return best
}
