package textrun

import (
	"sort"

	"github.com/chinmay-sawant/blinkless/layout"
)

// collect gathers the text runs inside box, grouped by baseline.
func collect(d *layout.Display, box layout.Box) []line {
	pt := d.PointsPerPixel
	left, right := box.X*pt, (box.X+box.W)*pt
	top, bottom := box.Y*pt, (box.Y+box.H)*pt

	var lines []line
	for i := range d.Ops {
		op := &d.Ops[i]
		if op.Kind != layout.DisplayOpText || op.Text == "" || op.Font == nil {
			continue
		}
		if op.Y < top-1 || op.Y > bottom+1 {
			continue
		}
		if op.X+op.W < left-1 || op.X > right+1 {
			continue
		}

		lines = addRun(lines, op)
	}

	sort.SliceStable(lines, func(i, j int) bool {
		return lines[i].baseline < lines[j].baseline
	})

	return lines
}

func addRun(lines []line, op *layout.DisplayOp) []line {
	for i := range lines {
		if lines[i].baseline < op.Y-0.6 || lines[i].baseline > op.Y+0.6 {
			continue
		}

		lines[i].runs = append(lines[i].runs, run{op: op})
		sort.SliceStable(lines[i].runs, func(a, b int) bool {
			return lines[i].runs[a].op.X < lines[i].runs[b].op.X
		})

		return lines
	}

	return append(lines, line{baseline: op.Y, runs: []run{{op: op}}})
}

// nearest returns the index of the line closest to a baseline.
func nearest(lines []line, yPt float64) int {
	best := 0
	dist := abs(lines[0].baseline - yPt)
	for i, ln := range lines[1:] {
		if d := abs(ln.baseline - yPt); d < dist {
			best, dist = i+1, d
		}
	}

	return best
}

// lineText joins the drawn text of a line's runs in reading order.
func lineText(ln line) string {
	out := ""
	for _, r := range ln.runs {
		out += drawn(r.op)
	}

	return out
}

func drawn(op *layout.DisplayOp) string {
	return layout.DisplayTransformText(op.Text, op.TextTransformValue())
}

func abs(v float64) float64 {
	if v < 0 {
		return -v
	}

	return v
}
