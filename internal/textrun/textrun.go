// Package textrun measures a point against the shaped text runs of a display
// list. A page uses it to turn a click into a rune offset.
package textrun

import (
	"github.com/chinmay-sawant/blinkless/layout"
)

// line is one baseline and the runs painted on it.
type line struct {
	baseline float64
	runs     []run
}

// run is one text operation.
type run struct {
	op *layout.DisplayOp
}

// At returns the rune offset in target nearest to (x, y) in CSS pixels.
// The runs are matched to target in reading order, allowing whitespace
// skipped at a line break. ok is false when the box holds no measurable run
// or the runs do not line up with target.
func At(d *layout.Display, box layout.Box, x, y float64, target string) (int, bool) {
	if d == nil || d.PointsPerPixel <= 0 {
		return 0, false
	}

	lines := collect(d, box)
	if len(lines) == 0 {
		return 0, false
	}

	bases, ok := align(lines, target)
	if !ok {
		return 0, false
	}

	pt := d.PointsPerPixel
	at := nearest(lines, y*pt)
	off, ok := offsetInLine(lines[at], x, 1/pt)
	if !ok {
		return 0, false
	}

	return bases[at] + off, true
}
