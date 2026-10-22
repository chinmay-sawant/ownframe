package main

import (
	"fmt"

	"github.com/chinmay-sawant/ownframe"
)

// catalogMetrics records the full-page row geometry once. The scroll window
// needs row tops and heights that are not present in a windowed layout.
type catalogMetrics struct {
	rowTop     []int
	rowH       []int
	gap        int
	gridTop    int
	gridBottom int
	contentH   int
}

// Grid shape the measure and windowing math share.
const (
	// cardsPerRow is the catalog grid width in cards.
	cardsPerRow = 2
	// maxCoord is larger than any CSS pixel coordinate the page reports.
	maxCoord = 1 << 30
)

// measureCatalog reads the full-page boxes after the two setup redraws.
func measureCatalog(p *ownframe.Page) (*catalogMetrics, bool) {
	boxes := p.Boxes()
	grid, ok := findBox(boxes, "catalog-grid")
	if !ok {
		return nil, false
	}
	m := &catalogMetrics{gridTop: int(grid.Y), gridBottom: int(grid.Y + grid.H)}
	for _, b := range boxes {
		if bottom := int(b.Y + b.H); bottom > m.contentH {
			m.contentH = bottom
		}
	}

	cards := (len(catalogProperties) + cardsPerRow - 1) / cardsPerRow
	for row := 0; row < cards; row++ {
		top, bottom := maxCoord, 0
		for col := 0; col < cardsPerRow; col++ {
			card := row*cardsPerRow + col
			if card >= len(catalogProperties) {
				break
			}
			b, ok := findBox(boxes, fmt.Sprintf("catalog-row-%02d", card+1))
			if !ok {
				return nil, false
			}
			if y := int(b.Y); y < top {
				top = y
			}
			if y := int(b.Y + b.H); y > bottom {
				bottom = y
			}
		}
		m.rowTop = append(m.rowTop, top)
		m.rowH = append(m.rowH, bottom-top)
	}

	if len(m.rowH) > 1 {
		m.gap = m.rowTop[1] - (m.rowTop[0] + m.rowH[0])
	}
	if m.gap <= 0 {
		m.gap = 8
	}

	return m, true
}
