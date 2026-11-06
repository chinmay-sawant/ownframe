package page

import (
	"image"
	"sync"

	"github.com/chinmay-sawant/ownframe/internal/page/dirty"
	"github.com/chinmay-sawant/ownframe/internal/page/hit"
)

// kidCaches holds one parent-to-children box index per page, rebuilt by
// setBoxes on every commit. A per-page entry is always fresh: setBoxes
// is the only writer of p.boxes and p.index, and it refreshes the entry
// in the same call, so any mark that finds boxes through the index
// finds the index built from those same boxes. Pages have no teardown
// hook, so entries live until the process exits; the proper home is a
// field on Page, which needs a wider change.
//
//nolint:gochecknoglobals // reason: process-wide per-page kid index keyed by page pointer, guarded by its own mutex; entries are refreshed wholesale by setBoxes, never mutated in place.
var kidCaches = struct {
	sync.Mutex
	byPage map[*Page]dirty.KidIndex
}{byPage: map[*Page]dirty.KidIndex{}}

// setBoxes stores the hit-test boxes from a Redraw. It rebuilds the id
// index so lookups stay O(1) on large pages, incrementally maintains
// the cached content rect, and rebuilds the parent-to-children index,
// so marks walk only the changed element's subtree. An unchanged id
// set repoints the index without clearing it, so a redraw of the same
// page keeps the map's capacity and buckets.
func (p *Page) setBoxes(boxes []Box) {
	old := p.boxes
	w, h := p.canvasSize()
	// Delta the content rect before the index is rebuilt in place: p.index
	// still describes old here, and Index clears it.
	content, ok := p.extendContent(p.index, boxes, w, h)
	p.boxes = boxes
	if hit.SameIDs(old, boxes) {
		p.index = hit.Repoint(p.index, boxes)
	} else {
		p.index = hit.Index(p.index, boxes)
	}
	p.hits = nil
	p.finishContent(old, content, ok, w, h)
	refreshKidIndex(p, boxes)
}

// canvasSize is the canvas the content rect covers: the committed
// display's size, or the window size before the first commit.
func (p *Page) canvasSize() (int, int) {
	if p.draw.Display != nil {
		return p.draw.Display.Width, p.draw.Display.Height
	}
	return p.width, p.height
}

// extendContent unions the new and grown boxes into the cached content
// rect, before the id index is rebuilt. ok is false when the cache is
// not current and the caller must recompute it fully: a cold cache, a
// resized canvas, a shrink, a move, or a missing old index. A pure
// grow stays incremental: the old rect sits inside the new one, so
// unioning the new rect keeps the union exact.
func (p *Page) extendContent(oldIndex map[string]*Box, boxes []Box, w, h int) (image.Rectangle, bool) {
	if !p.contentOK || oldIndex == nil || p.contentW != w || p.contentH != h {
		return image.Rectangle{}, false
	}
	content := p.content
	for _, b := range boxes {
		r := boxRect(b)
		ob, found := oldIndex[b.ID]
		if !found {
			content = content.Union(r)
			continue
		}
		if or := boxRect(*ob); r != or {
			if !or.In(r) {
				return content, false
			}
			content = content.Union(r)
		}
	}
	return content, true
}

// finishContent completes the incremental content-rect maintenance after
// the id index is rebuilt. A box removed since the last commit may
// have carried the union's edge, and the union cannot shrink
// incrementally, so any removal recomputes the canvas plus the current
// boxes. Duplicate ids that lose one copy keep the survivor's rect
// covered: the union over-covers rather than recomputes, which
// repaints more, never less.
func (p *Page) finishContent(old []Box, content image.Rectangle, ok bool, w, h int) {
	if ok {
		for i := range old {
			if _, found := p.index[old[i].ID]; !found {
				ok = false
				break
			}
		}
	}
	if !ok {
		content = image.Rect(0, 0, w, h).Union(unionBoxes(p.boxes))
	}
	p.content, p.contentW, p.contentH, p.contentOK = content, w, h, true
}

// unionBoxes returns the union of every box rect.
func unionBoxes(boxes []Box) image.Rectangle {
	r := image.Rectangle{}
	for _, b := range boxes {
		r = r.Union(boxRect(b))
	}
	return r
}

// refreshKidIndex rebuilds the page's parent-to-children box index from
// the committed boxes.
func refreshKidIndex(p *Page, boxes []Box) {
	idx := dirty.BuildKidIndex(boxes)
	kidCaches.Lock()
	kidCaches.byPage[p] = idx
	kidCaches.Unlock()
}

// kidRect returns the union of the id's box and its descendants' boxes
// from the page's parent-to-children index. ok is false when the index
// has no entry for the page or the id, and the caller falls back to the
// id's own box.
func kidRect(p *Page, id string) (image.Rectangle, bool) {
	kidCaches.Lock()
	idx, found := kidCaches.byPage[p]
	kidCaches.Unlock()
	if !found {
		return image.Rectangle{}, false
	}
	return idx.Rect(id)
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
// rect is cached per canvas size until the next setBoxes, so the
// per-frame TakeDirty costs O(1) on large pages. setBoxes maintains it
// incrementally; the size check stays as the backstop for restores that
// bypass setBoxes.
func (p *Page) contentRect() image.Rectangle {
	w, h := p.canvasSize()
	if p.contentOK && p.contentW == w && p.contentH == h {
		return p.content
	}
	r := image.Rect(0, 0, w, h).Union(unionBoxes(p.boxes))
	p.content, p.contentW, p.contentH, p.contentOK = r, w, h, true
	return r
}
