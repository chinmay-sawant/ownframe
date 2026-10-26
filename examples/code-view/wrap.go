package main

import (
	"sort"
	"strings"
)

const (
	wrapCellEstimate = 8
	wrapFloorDiv     = 3
	wrapFloorMin     = 4
	segmentEstimate  = 4
	planPair         = 2
)

// wrapPlan maps source lines onto fixed-height visual rows. A plan is cached
// on the document and rebuilt only when the available column count changes.
type wrapPlan struct {
	Cols  int
	Rows  []codeRow
	First []int
}

// wrapCell is one rune with its token class.
type wrapCell struct {
	r     rune
	class string
}

// wrapped returns the plan for cols, reusing the previous plan on a hit.
func (d *fileDoc) wrapped(cols int) *wrapPlan {
	if cols < 1 {
		cols = 1
	}

	if d.Wrap != nil && d.Wrap.Cols == cols {
		return d.Wrap
	}

	plan := &wrapPlan{Cols: cols, First: make([]int, len(d.Tokens)+1)}

	for line, tokens := range d.Tokens {
		plan.First[line] = len(plan.Rows)
		plan.wrapLine(line, tokens, cols)
	}

	plan.First[len(d.Tokens)] = len(plan.Rows)
	d.Wrap = plan

	return plan
}

// wrapLine appends the visual rows for one source line. It prefers a break at
// a space and falls back to a hard edge break for one long token. A
// continuation row has No zero.
func (p *wrapPlan) wrapLine(line int, tokens []segment, cols int) {
	cells := make([]wrapCell, 0, len(tokens)*wrapCellEstimate)

	for _, token := range tokens {
		for _, r := range token.Text {
			cells = append(cells, wrapCell{r: r, class: token.Class})
		}
	}

	if len(cells) == 0 {
		p.Rows = append(p.Rows, codeRow{No: line + 1, Line: line})
		return
	}

	start := 0
	first := true

	for len(cells) > 0 {
		cut := cols
		if cut > len(cells) {
			cut = len(cells)
		}

		if len(cells) > cols {
			// Leave at least a few cells on the row before looking for a
			// space, so indentation does not force an almost empty row.
			floor := cols / wrapFloorDiv
			if floor < wrapFloorMin {
				floor = 4
			}

			for i := cut - 1; i >= floor; i-- {
				if cells[i].r == ' ' {
					cut = i + 1

					break
				}
			}
		}

		row := codeRow{Line: line, Start: start, Segments: cellsToSegments(cells[:cut])}
		if first {
			row.No = line + 1
			first = false
		}

		p.Rows = append(p.Rows, row)
		cells = cells[cut:]
		start += cut
	}
}

// cellsToSegments merges adjacent runes of the same class.
func cellsToSegments(cells []wrapCell) []segment {
	out := make([]segment, 0, segmentEstimate)

	var (
		run   strings.Builder
		class string
	)

	flush := func() {
		if run.Len() == 0 {
			return
		}

		out = append(out, segment{Text: run.String(), Class: class})
		run.Reset()
	}

	for _, cell := range cells {
		if class != cell.class {
			flush()
			class = cell.class
		}

		run.WriteRune(cell.r)
	}

	flush()

	return out
}

// firstRow returns the visual row that starts one source line.
func (p *wrapPlan) firstRow(line int) int {
	if line < 0 {
		return 0
	}

	if line >= len(p.First)-1 {
		return p.First[len(p.First)-1]
	}

	return p.First[line]
}

// lineAt returns the source line that owns one visual row.
func (p *wrapPlan) lineAt(row int) int {
	if len(p.First) < planPair {
		return 0
	}

	line := sort.Search(len(p.First)-1, func(i int) bool {
		return p.First[i+1] > row
	})

	if line >= len(p.First)-1 {
		return len(p.First) - planPair
	}

	return line
}
