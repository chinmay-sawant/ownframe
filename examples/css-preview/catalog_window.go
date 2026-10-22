package main

import (
	"context"

	"github.com/chinmay-sawant/ownframe"
)

// catalogGridID is the element that holds the rows. One patch of the grid
// carries the state of every row inside it.
const catalogGridID = "catalog-grid"

// catalogWindow slices the card list into the rows the viewport shows. The
// spacers keep the grid height and the scroll range at the full-page values.
type catalogWindow struct {
	prevStart int
	prevEnd   int
	metrics   *catalogMetrics
	view      *view
	demo      *demo
	page      *ownframe.Page
}

// installCatalogWindow replaces the full card list with the rows the
// viewport shows. The spacers keep the grid height and the scroll range at
// the full-page values.
func installCatalogWindow(
	ctx context.Context, p *ownframe.Page, v *view, d *demo, m *catalogMetrics,
) error {
	w := &catalogWindow{prevStart: -1, prevEnd: -1, metrics: m, view: v, demo: d, page: p}
	d.window = w
	p.SetWindowing(true)
	p.SetScrollWindow(w.slice)
	p.SetScrollOffset(0, 0)

	_, viewH := p.Size()
	w.slice(0, viewH)
	p.SetData(v)

	return p.Redraw(ctx)
}

// slice puts the rows the viewport shows into the view. It reports whether
// the window moved, so the page skips the layout when it did not.
func (w *catalogWindow) slice(offsetY, viewH int) bool {
	w.demo.checkSize()
	m := w.metrics
	start, end := m.rangeFor(offsetY, viewH)
	if start == w.prevStart && end == w.prevEnd {
		return false
	}
	w.prevStart, w.prevEnd = start, end

	v := w.view
	cardStart, cardEnd := start*cardsPerRow, end*cardsPerRow
	if cardEnd > len(v.Catalog) {
		cardEnd = len(v.Catalog)
	}
	v.Rows = v.Catalog[cardStart:cardEnd]
	v.TopPad = m.topPad(start)
	v.BotPad = m.botPad(end)
	w.page.SetData(v)
	// A row that scrolled into view starts the preview it still owes.
	w.demo.queueRows(v.Rows)
	w.demo.wake()

	return true
}
