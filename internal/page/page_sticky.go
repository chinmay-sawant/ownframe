package page

import (
	"sort"

	"github.com/chinmay-sawant/blinkless/layout"
	"github.com/chinmay-sawant/ownframe/internal/page/hit"
)

type scrollViewport struct{ x, y, width, height float64 }

// SetScrollViewport is the native host's logical scroll-space handoff. Public
// callers keep using ScrollTo/ScrollBy; this consumes the host's actual clamp
// and zoom after those requests have been applied.
func (p *Page) SetScrollViewport(x, y, width, height float64) {
	if !p.hasViewportSticky {
		p.stickyViewport = nil
		return
	}
	next := scrollViewport{x: x, y: y, width: width, height: height}
	if p.stickyViewport != nil && *p.stickyViewport == next {
		return
	}
	p.stickyViewport = &next
	p.refreshStickyBoxes(x, y, width, height)
}

func (p *Page) refreshStickyBoxes(x, y, width, height float64) {
	display := p.Display()
	if !p.hasViewportSticky || display == nil {
		return
	}
	p.setStickyBoxes(layout.StickyView(display, x, y, width, height))
}

func (p *Page) setDisplayBoxes(display *Display) {
	p.hasViewportSticky = layout.HasViewportSticky(display)
	if !p.hasViewportSticky {
		p.stickyViewport = nil
		p.setPlacedBoxes(display)
		return
	}
	viewport := scrollViewport{x: float64(p.scroll.Windowing.X), y: float64(p.scroll.Windowing.Y), width: float64(p.width), height: float64(p.height)}
	if p.stickyViewport != nil {
		viewport = *p.stickyViewport
	}
	p.setStickyBoxes(layout.StickyView(display, viewport.x, viewport.y, viewport.width, viewport.height))
}

// A sticky stacking context can paint above a later in-flow sibling. Preserve
// document order in Boxes, but route pointer input in the engine's paint order.
func (p *Page) setStickyBoxes(display *Display) {
	p.setPlacedBoxes(display)
	items := hit.List(nil, display.Boxes, fixedBoxes(display))
	ranks := make([]int, len(items))
	for rank, operation := range display.Order {
		if operation < 0 || operation >= len(display.OpOwners) || display.Ops[operation].Outline() {
			continue
		}
		owner := display.OpOwners[operation]
		if owner >= 0 && owner < len(ranks) {
			ranks[owner] = rank + 1
		}
	}
	for index, box := range display.Boxes {
		if ranks[index] == 0 && box.ParentIndex >= 0 && box.ParentIndex < index {
			ranks[index] = ranks[box.ParentIndex]
		}
	}
	order := make([]int, len(items))
	for index := range order {
		order[index] = index
	}
	sort.SliceStable(order, func(a, b int) bool { return ranks[order[a]] < ranks[order[b]] })
	p.hits = make([]hitItem, len(items))
	for index, original := range order {
		p.hits[index] = items[original]
	}
}
