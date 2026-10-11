package textrun

import (
	"unicode/utf8"

	"github.com/chinmay-sawant/blinkless/layout"
	"github.com/hajimehoshi/ebiten/v2/text/v2"
)

// Caret returns the text style and x coordinate for a rune offset. x uses
// display points, like the returned operation's baseline.
func Caret(d *layout.Display, box layout.Box, target string, offset int) (layout.DisplayOp, float64, bool) {
	if d == nil || d.PixelPerPoint <= 0 {
		return layout.DisplayOp{}, 0, false
	}
	lines := collect(d, box)
	if len(lines) == 0 {
		return layout.DisplayOp{}, 0, false
	}
	if target == "" {
		op := lines[0].runs[0].op

		return *op, op.X, true
	}
	bases, ok := align(lines, target)
	if !ok {
		return layout.DisplayOp{}, 0, false
	}
	offset = max(0, min(offset, utf8.RuneCountInString(target)))
	for i, ln := range lines {
		local := offset - bases[i]
		if local < 0 || local > utf8.RuneCountInString(lineText(ln)) {
			continue
		}
		for _, run := range ln.runs {
			op := run.op
			n := utf8.RuneCountInString(drawn(op))
			if local <= n {
				dx, ok := advance(op, local, d.PixelPerPoint)
				return *op, op.X + dx*d.PointsPerPixel, ok
			}
			local -= n
		}
	}
	return layout.DisplayOp{}, 0, false
}

func advance(op *layout.DisplayOp, count int, pxPerPt float64) (float64, bool) {
	runes := []rune(drawn(op))
	count = max(0, min(count, len(runes)))
	face := faceFor(op, pxPerPt)
	if face == nil {
		return 0, false
	}
	if op.LetterSpacing == 0 {
		width, _ := text.Measure(string(runes[:count]), face, 0)
		return width, true
	}
	var width float64
	for _, r := range runes[:count] {
		part, _ := text.Measure(string(r), face, 0)
		width += part + op.LetterSpacing*pxPerPt
	}
	return width, true
}
