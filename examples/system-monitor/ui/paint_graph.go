package ui

import (
	"github.com/chinmay-sawant/ownframe"
	"github.com/chinmay-sawant/ownframe/internal/frame"
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

	bx, by, bw, bh := frame.BoxUnits(d, box)
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
	minH := 1.5 * pp
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

// graphColumns picks a column count from the graph's visible width, one bar
// per four CSS pixels, bounded by the bar nodes the template rendered.
func graphColumns(width, pixelPerPoint float64) int {
	cols := int(width / (4 * pixelPerPoint))

	return min(max(cols, 24), graphColsMax)
}
