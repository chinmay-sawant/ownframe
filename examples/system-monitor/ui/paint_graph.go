package ui

import (
	"github.com/chinmay-sawant/ownframe"
)

// paintGraph lays every bar of one card from the downsampled ring. The bar
// nodes come from the template; the tick sets their geometry, so the graph
// spans the visible width at one bar per few pixels and never renders more
// bars than the width can show.
func (a *App) paintGraph(d *ownframe.Display, p *panel) {
	bars := a.state.h.bars[p.id]
	box, ok := a.state.h.graph[p.id]
	if !ok || len(bars) == 0 {
		return
	}

	bx, by, bw, bh := ownframe.BoxUnits(d, box)
	bottom := by + bh
	pp := d.PointsPerPixel

	cols := min(graphColumns(bw, pp), len(bars))
	vals, oks := p.graph.Columns(cols)

	scale := p.last.Scale
	if scale <= 0 {
		if m, ok := p.graph.Max(); ok {
			scale = m
		}
	}

	pitch := bw / float64(cols)
	barW := max(pitch-pp, pp)
	minH := barMinHeight * pp
	avail := max(bh-minH, 0)

	for i, bar := range bars {
		if i >= cols {
			bar.W, bar.H = 0, 0

			continue
		}

		frac := 0.0

		if oks[i] && scale > 0 {
			frac = min(max(vals[i]/scale, 0), 1)
		}

		bar.X = bx + float64(i)*pitch
		bar.W = barW
		bar.H = minH + frac*avail
		bar.Y = bottom - bar.H
	}
}

// Graph geometry: the minimum bar height in points, the CSS pixels per
// bar, and the column clamp.
const (
	barMinHeight = 1.5
	pxPerBar     = 4
	graphColsMin = 24
)

// graphColumns picks a column count from the graph's visible width, one bar
// per few CSS pixels, bounded by the bar nodes the template rendered.
func graphColumns(width, pixelPerPoint float64) int {
	cols := int(width / (pxPerBar * pixelPerPoint))

	return min(max(cols, graphColsMin), graphColsMax)
}
