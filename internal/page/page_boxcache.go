package page

import (
	"image"

	"github.com/chinmay-sawant/ownframe/internal/page/hit"
)

// setBoxes stores the hit-test boxes from a Redraw. It rebuilds the id
// index so lookups stay O(1) on large pages, and drops the cached content
// rect. Boxes change only here, so the index never goes stale.
//
// An unchanged id set repoints the index without clearing it, so a redraw of
// the same page keeps the map's capacity and buckets. An unchanged box list
// keeps the cached content rect: the rect is a pure function of the boxes
// and the canvas size, and the size check on read stays as the backstop.
func (p *Page) setBoxes(boxes []Box) {
	old := p.boxes
	p.boxes = boxes
	if hit.SameIDs(old, boxes) {
		p.index = hit.Repoint(p.index, boxes)
	} else {
		p.index = hit.Index(p.index, boxes)
	}
	p.hits = nil
	if !sameBoxSlice(old, boxes) {
		p.contentOK = false
	}
}

// sameBoxSlice reports whether two box lists name the same elements with the
// same rectangles, in order. It reuses the trigger box identity: the id, the
// tag, the parent link, and the rect are exactly what the content rect and
// the hit list derive from.
func sameBoxSlice(a, b []Box) bool {
	if len(a) != len(b) {
		return false
	}
	for i := range a {
		if !sameBox(a[i], b[i]) {
			return false
		}
	}
	return true
}

func (p *Page) setDisplayBoxes(d *Display) {
	fixed := fixedBoxes(d)
	if len(fixed) == 0 {
		p.setBoxes(d.Boxes)
		p.hits = nil

		return
	}
	if p.hitsReusable(d.Boxes, fixed) {
		// Same elements, same rectangles, same pins: the stored list already
		// describes the new display, so keep it and its capacity instead of
		// recopying every item. The index still repoints at the new boxes.
		keep := p.hits
		p.setBoxes(d.Boxes)
		p.hits = keep

		return
	}
	hits := p.hits
	p.setBoxes(d.Boxes)
	p.hits = hit.List(hits, d.Boxes, fixed)
}

// hitsReusable reports whether the stored hit list already describes boxes
// with the same pins: the same elements with the same rectangles in order,
// and the same fixed roots as sets. The markings are a pure function of the
// box links and the roots, so equal inputs mean an identical list.
func (p *Page) hitsReusable(boxes, fixed []Box) bool {
	if len(p.hits) != len(boxes) {
		return false
	}
	for i := range boxes {
		if !sameBox(p.hits[i].Box, boxes[i]) {
			return false
		}
	}
	return sameFixedRoots(p.hits, fixed)
}

// sameFixedRoots reports whether the fixed roots behind hits equal fixed as
// sets. Roots are the fixed items whose parent is not fixed: fixedness only
// propagates down from a root, so a fixed item with a non-fixed parent is a
// root itself, and every root qualifies. The greedy match is exact because
// box identity is an equivalence: within one class any pairing is correct,
// across classes no pairing matches.
func sameFixedRoots(hits []hitItem, fixed []Box) bool {
	roots := make([]Box, 0, len(fixed))
	for i := range hits {
		if !hits[i].Fixed {
			continue
		}
		if pi := hits[i].Box.ParentIndex; pi >= 0 && pi < len(hits) && hits[pi].Fixed {
			continue
		}
		roots = append(roots, hits[i].Box)
	}
	if len(roots) != len(fixed) {
		return false
	}
	used := make([]bool, len(roots))
	for _, fb := range fixed {
		found := false
		for j := range roots {
			if !used[j] && sameBox(roots[j], fb) {
				used[j] = true
				found = true
				break
			}
		}
		if !found {
			return false
		}
	}
	return true
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
