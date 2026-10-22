package page

import "github.com/chinmay-sawant/blinkless/layout"

// spliceElement adopts the patched page. The engine hands over the whole
// display, because a patch that changes the operation count moves every later
// operation index. The page rebuilds its box index, marks the changed
// operations dirty, drops the picture, and refreshes the a11y tree.
func (p *Page) spliceElement(patch *layout.ElementPatch, id string, edits []ElementEdit) bool {
	if patch == nil || patch.Display == nil {
		return false
	}
	d := patch.Display
	prev := p.draw.SwapList(d)
	p.setBoxes(d.Boxes)
	p.hits = hitList(d.Boxes, fixedBoxes(d))
	p.dirtyFromDisplay(prev, d)
	p.generation++
	p.stats.layouts++
	p.stats.repaints++
	p.refreshAccessibility(id, edits)

	return true
}

// rollbackPatch undoes the edits of a refused patch. The cached tree is
// dropped, so the next redraw parses the source again and the page shows the
// source's state. The render key is forgotten, so it redraws.
func (p *Page) rollbackPatch() {
	p.invalidateCache()
	p.hasLast = false
}
