package main

import (
	"context"

	"github.com/chinmay-sawant/ownframe"
)

// installCatalogWindow replaces the full card list with the visible rows and
// two rows of overscan. The spacers keep the grid height and scroll range at
// the full-page values.
func installCatalogWindow(
	ctx context.Context, p *ownframe.Page, v *view, d *demo, m *catalogMetrics,
) error {
	prevStart, prevEnd := -1, -1

	apply := func(offsetY, viewH int) bool {
		if d.allPlaying {
			return false
		}
		start, end := m.rangeFor(offsetY, viewH)
		if start == prevStart && end == prevEnd {
			return false
		}
		prevStart, prevEnd = start, end

		cardStart, cardEnd := start*2, end*2
		if cardEnd > len(v.Catalog) {
			cardEnd = len(v.Catalog)
		}
		v.Rows = v.Catalog[cardStart:cardEnd]
		v.TopPad = m.topPad(start)
		v.BotPad = m.botPad(end)
		p.SetData(v)

		return true
	}

	p.SetWindowing(true)
	p.SetScrollWindow(apply)
	p.SetScrollOffset(0, 0)

	_, viewH := p.Size()
	apply(0, viewH)
	p.SetData(v)

	return p.Redraw(ctx)
}
