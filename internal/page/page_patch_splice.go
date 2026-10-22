package page

import (
	"strings"

	"github.com/chinmay-sawant/blinkless/layout"
	"github.com/chinmay-sawant/blinkless/markup"
)

// spliceElement adopts the patched page. The engine hands over the whole
// display, because a patch that changes the operation count moves every later
// operation index. The page rebuilds its box index, marks the changed
// operations dirty, drops the picture, and refreshes the a11y tree.
func (p *Page) spliceElement(patch *layout.ElementPatch, id string, edits []ElementEdit) bool {
	if patch == nil || patch.Display == nil {
		return false
	}
	prev := p.display
	p.display = patch.Display
	p.setBoxes(p.display.Boxes)
	p.hits = hitList(p.display.Boxes, fixedBoxes(p.display))
	p.dirtyFromDisplay(prev, p.display)
	p.png = nil
	p.generation++
	p.stats.layouts++
	p.stats.repaints++
	p.refreshAccessibility(id, edits)

	return true
}

// refreshAccessibility puts the patch's edits into the accessibility tree and
// retires the document, so an old tree is not served as current. A tree no
// client asked for yet is built from the source when one does.
func (p *Page) refreshAccessibility(id string, edits []ElementEdit) {
	s := p.a11yState()
	if s == nil {
		return
	}
	if s.root != nil {
		b := newAccessibilityBuilder(p, s.root)
		for _, e := range edits {
			if e.ID == "" {
				e.ID = id
			}
			for _, n := range b.ids[e.ID] {
				if e.Text != nil {
					n.Children = []*markup.Node{{Type: markup.TypeText, Text: *e.Text}}
				}
				if e.Attribute != "" {
					if n.Attrs == nil {
						n.Attrs = map[string]string{}
					}
					n.Attrs[strings.ToLower(e.Attribute)] = e.Value
				}
			}
		}
	}
	p.invalidateAccessibilityDocument()
	p.commitAccessibility(s.source)
}

// rollbackPatch undoes the edits of a refused patch. The cached tree is
// dropped, so the next redraw parses the source again and the page shows the
// source's state. The render key is forgotten, so it redraws.
func (p *Page) rollbackPatch() {
	p.invalidateCache()
	p.hasLast = false
}
