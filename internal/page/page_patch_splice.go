package page

import (
	"github.com/chinmay-sawant/blinkless/layout"
)

// spliceElement adopts the patched page. The engine hands over the whole
// display, because a patch that changes the operation count moves every later
// operation index, and with it the paint order and the owners. The page
// rebuilds its box index and marks the changed region dirty.
func (p *Page) spliceElement(patch *layout.ElementPatch) bool {
	if patch == nil || patch.Display == nil {
		return false
	}
	p.display = patch.Display
	println("ZZDEBUG splice boxes", len(p.display.Boxes), len(p.display.Ops))
	p.setBoxes(p.display.Boxes)
	p.hits = hitList(p.display.Boxes, fixedBoxes(p.display))
	p.markRect(patch.Old.Union(patch.New))
	p.generation++
	p.stats.layouts++
	p.stats.repaints++

	return true
}
