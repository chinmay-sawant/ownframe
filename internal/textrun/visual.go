package textrun

import (
	"unicode/utf8"

	"github.com/chinmay-sawant/blinkless/layout"
)

const (
	defaultLineHeightRatio = 1.25
	defaultFallbackSize    = 12.0
	emptyCaretXOffset      = 7.0
	emptyCaretTopPad       = 4.0
	lineCenterAscentRatio  = 0.4
)

// vline is one visual line in a field: either a painted line of text runs or
// an empty line created by a leading, consecutive, or trailing newline.
type vline struct {
	start    int
	end      int
	baseline float64
	leftX    float64
	sample   layout.DisplayOp
	painted  *line
}

// visualLines maps target onto the laid-out visual lines inside box, including
// empty lines produced by leading, consecutive, or trailing '\n' runes.
func visualLines(d *layout.Display, box layout.Box, target string) ([]vline, bool) {
	if d == nil || d.PointsPerPixel <= 0 {
		return nil, false
	}

	lines := collect(d, box)
	value := []rune(target)
	if len(lines) == 0 {
		return emptyVisualLines(d, box, value)
	}

	bases, ok := align(lines, target)
	if !ok {
		return nil, false
	}

	sample := *lines[0].runs[0].op
	leftX := sample.X
	step := measureLineStep(lines, bases, value, sample.Size)

	var out []vline

	// Leading empty lines before the first painted line.
	if leadNLs := newlineIndices(value, 0, bases[0]); len(leadNLs) > 0 {
		for j, nl := range leadNLs {
			s := 0
			if j > 0 {
				s = leadNLs[j-1] + 1
			}
			b := lines[0].baseline - float64(len(leadNLs)-j)*step
			out = append(out, vline{
				start:    s,
				end:      nl,
				baseline: b,
				leftX:    leftX,
				sample:   sample,
			})
		}
	}

	for i := range lines {
		ln := &lines[i]
		start := bases[i]
		end := start + utf8.RuneCountInString(lineText(*ln))
		out = append(out, vline{
			start:    start,
			end:      end,
			baseline: ln.baseline,
			leftX:    ln.runs[0].op.X,
			sample:   *ln.runs[0].op,
			painted:  ln,
		})

		if i+1 < len(lines) {
			midNLs := newlineIndices(value, end, bases[i+1])
			if len(midNLs) > 1 {
				dy := lines[i+1].baseline - ln.baseline
				for j := 1; j < len(midNLs); j++ {
					s := midNLs[j-1] + 1
					e := midNLs[j]
					b := ln.baseline + dy*float64(j)/float64(len(midNLs))
					out = append(out, vline{
						start:    s,
						end:      e,
						baseline: b,
						leftX:    leftX,
						sample:   sample,
					})
				}
			}
		}
	}

	// Trailing empty lines after the last painted line.
	lastIdx := len(lines) - 1
	lastEnd := bases[lastIdx] + utf8.RuneCountInString(lineText(lines[lastIdx]))
	if trailNLs := newlineIndices(value, lastEnd, len(value)); len(trailNLs) > 0 {
		lastBase := lines[lastIdx].baseline
		for j, nl := range trailNLs {
			s := nl + 1
			e := len(value)
			if j+1 < len(trailNLs) {
				e = trailNLs[j+1]
			}
			out = append(out, vline{
				start:    s,
				end:      e,
				baseline: lastBase + float64(j+1)*step,
				leftX:    leftX,
				sample:   sample,
			})
		}
	}

	return out, true
}

func emptyVisualLines(d *layout.Display, box layout.Box, value []rune) ([]vline, bool) {
	if len(value) == 0 || !allSpace(value) {
		return nil, false
	}

	pt := d.PointsPerPixel
	size := defaultFallbackSize
	step := size * defaultLineHeightRatio
	leftX := (box.X + emptyCaretXOffset) * pt
	firstBase := (box.Y+emptyCaretTopPad)*pt + size
	sample := layout.DisplayOp{Y: firstBase, Size: size, Alpha: 1}

	nls := newlineIndices(value, 0, len(value))
	out := make([]vline, 0, len(nls)+1)
	start := 0
	for j, nl := range nls {
		out = append(out, vline{
			start:    start,
			end:      nl,
			baseline: firstBase + float64(j)*step,
			leftX:    leftX,
			sample:   sample,
		})
		start = nl + 1
	}
	out = append(out, vline{
		start:    start,
		end:      len(value),
		baseline: firstBase + float64(len(nls))*step,
		leftX:    leftX,
		sample:   sample,
	})

	return out, true
}

func measureLineStep(lines []line, bases []int, value []rune, sampleSize float64) float64 {
	for i := 1; i < len(lines); i++ {
		prevEnd := bases[i-1] + utf8.RuneCountInString(lineText(lines[i-1]))
		breaks := len(newlineIndices(value, prevEnd, bases[i]))
		steps := max(1, breaks)
		dy := (lines[i].baseline - lines[i-1].baseline) / float64(steps)
		if dy > 0 {
			return dy
		}
	}

	if sampleSize <= 0 {
		sampleSize = defaultFallbackSize
	}

	return sampleSize * defaultLineHeightRatio
}

func newlineIndices(value []rune, from, to int) []int {
	if from < 0 {
		from = 0
	}
	if to > len(value) {
		to = len(value)
	}
	var out []int
	for i := from; i < to; i++ {
		if value[i] == '\n' {
			out = append(out, i)
		}
	}

	return out
}

func findVLine(vlines []vline, offset int) int {
	best := 0
	for i, vl := range vlines {
		if offset >= vl.start && offset <= vl.end {
			best = i
		} else if offset > vl.end {
			best = i
		}
	}

	return best
}

func nearestVLine(vlines []vline, yPt float64) int {
	best := 0
	dist := abs(vlineCenter(vlines[0]) - yPt)
	for i := 1; i < len(vlines); i++ {
		if d := abs(vlineCenter(vlines[i]) - yPt); d < dist {
			best, dist = i, d
		}
	}

	return best
}

func vlineCenter(vl vline) float64 {
	if vl.sample.Size > 0 {
		return vl.baseline - vl.sample.Size*lineCenterAscentRatio
	}

	return vl.baseline
}

func vlineCaretX(d *layout.Display, vl vline, offset int) (layout.DisplayOp, float64, bool) {
	if vl.painted == nil {
		op := vl.sample
		op.Y = vl.baseline

		return op, vl.leftX, true
	}

	local := max(0, min(offset-vl.start, vl.end-vl.start))
	for _, r := range vl.painted.runs {
		op := r.op
		n := utf8.RuneCountInString(drawn(op))
		if local <= n {
			dx, ok := advance(op, local, d.PixelPerPoint)

			return *op, op.X + dx*d.PointsPerPixel, ok
		}
		local -= n
	}

	return layout.DisplayOp{}, 0, false
}

// MoveVertical moves offset by dir (-1 up, +1 down) visual lines inside box.
// wantX is the preferred x coordinate in display points; pass a negative value
// to measure it from offset. Returns the new rune offset, the preferred x to
// carry into the next consecutive vertical move, and whether layout lines were
// available.
func MoveVertical(d *layout.Display, box layout.Box, target string, offset, dir int, wantX float64) (int, float64, bool) {
	if d == nil || d.PixelPerPoint <= 0 || dir == 0 {
		return offset, wantX, false
	}

	vlines, ok := visualLines(d, box, target)
	if !ok || len(vlines) == 0 {
		return offset, wantX, false
	}

	n := utf8.RuneCountInString(target)
	offset = max(0, min(offset, n))
	cur := findVLine(vlines, offset)
	if wantX < 0 {
		_, x, ok := vlineCaretX(d, vlines[cur], offset)
		if !ok {
			return offset, wantX, false
		}
		wantX = x
	}

	next := cur + dir
	if next < 0 {
		return 0, wantX, true
	}
	if next >= len(vlines) {
		return n, wantX, true
	}

	vl := vlines[next]
	if vl.painted == nil {
		return vl.start, wantX, true
	}

	off, ok := offsetInLine(*vl.painted, wantX*d.PixelPerPoint, d.PixelPerPoint)
	if !ok {
		return offset, wantX, false
	}

	return vl.start + off, wantX, true
}

// LineRange returns the [start, end] rune offsets of the visual line containing
// offset inside box.
func LineRange(d *layout.Display, box layout.Box, target string, offset int) (int, int, bool) {
	if d == nil || d.PointsPerPixel <= 0 {
		return 0, 0, false
	}

	vlines, ok := visualLines(d, box, target)
	if !ok || len(vlines) == 0 {
		return 0, 0, false
	}

	n := utf8.RuneCountInString(target)
	offset = max(0, min(offset, n))
	vl := vlines[findVLine(vlines, offset)]

	return vl.start, vl.end, true
}
