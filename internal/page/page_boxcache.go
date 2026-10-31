package page

import (
	"image"

	"github.com/chinmay-sawant/ownframe/internal/page/hit"
)

// setBoxes stores the hit-test boxes from a Redraw. It rebuilds the id
// index so lookups stay O(1) on large pages, and drops the cached content
// rect. Boxes change only here, so the index never goes stale.
func (p *Page) setBoxes(boxes []Box) {
	p.boxes = boxes
	p.hits = nil
	p.index = hit.Index(p.index, boxes)
	p.contentOK = false
}

func (p *Page) setDisplayBoxes(d *Display) {
	hits := p.hits
	p.setBoxes(d.Boxes)
	fixed := fixedBoxes(d)
	if len(fixed) > 0 {
		p.hits = hit.List(hits, d.Boxes, fixed)
	} else {
		p.hits = nil
	}
}

// Boxes returns a copy of the last hit-test boxes, or nil when nothing has
// been drawn. Caller edits never corrupt the page index.
func (p *Page) Boxes() []Box {
	return append([]Box(nil), p.boxes...)
}

// boxByID returns the last box laid out with that id, or Box{ID: id}. The
// lookup is one map read, never a walk over p.boxes.
func (p *Page) boxByID(id string) Box {
	return hit.ByID(p.index, id)
}

// hasBox reports whether the last layout has a box for id, in one map read.
func (p *Page) hasBox(id string) bool {
	return hit.Has(p.index, id)
}

// contentRect is the painted page in CSS pixels: the canvas plus any box
// that overflows it, the region the window's replay buffer covers. The
// rect is cached per canvas size until the next setBoxes, so the per-frame
// TakeDirty costs O(1) on large pages.
func (p *Page) contentRect() image.Rectangle {
	w, h := p.width, p.height
	if p.draw.Display != nil {
		w, h = p.draw.Display.Width, p.draw.Display.Height
	}
	if p.contentOK && p.contentW == w && p.contentH == h {
		return p.content
	}
	r := image.Rect(0, 0, w, h)
	for _, b := range p.boxes {
		r = r.Union(boxRect(b))
	}
	p.content, p.contentW, p.contentH, p.contentOK = r, w, h, true
	return r
}
