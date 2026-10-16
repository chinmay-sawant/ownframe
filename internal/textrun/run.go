package textrun

import (
	"math"
	"unicode/utf8"
)

// offsetInLine returns the rune offset nearest xPx among the line's runs.
// xPx is a CSS pixel distance from the display's left edge. ok is false when
// a run's font cannot be measured.
func offsetInLine(ln line, xPx, pxPerPt float64) (int, bool) {
	best, bestBase := 0, 0
	bestDist := math.MaxFloat64
	base := 0
	for i, r := range ln.runs {
		if faceFor(r.op, pxPerPt) == nil {
			return 0, false
		}

		left := r.op.X * pxPerPt
		right := left + runWidth(r.op, pxPerPt)
		if xPx >= left && xPx <= right {
			return base + offsetIn(r.op, xPx-left, pxPerPt), true
		}

		dist := left - xPx
		if xPx > right {
			dist = xPx - right
		}
		if dist < bestDist {
			best, bestBase, bestDist = i, base, dist
		}

		base += utf8.RuneCountInString(drawn(r.op))
	}

	op := ln.runs[best].op

	return bestBase + offsetIn(op, xPx-op.X*pxPerPt, pxPerPt), true
}
