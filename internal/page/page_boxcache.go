package page

import (
	"image"
	"sync"

	"github.com/chinmay-sawant/ownframe/internal/page/dirty"
	"github.com/chinmay-sawant/ownframe/internal/page/hit"
)

// boxCaches holds one pageBoxes per page, keyed by page pointer and guarded
// by its own mutex. setBoxes publishes into the entry; kidRect and
// displayDeclarations read it. Pages have no teardown hook, so entries
// live until the process exits; the proper home is a field on Page, which
// needs a wider change.
//
//nolint:gochecknoglobals // reason: process-wide per-page box cache keyed by page pointer, guarded by its own mutex; entries are replaced wholesale by setBoxes, never mutated from another page.
var boxCaches = struct {
	sync.Mutex
	byPage map[*Page]*pageBoxes
}{byPage: map[*Page]*pageBoxes{}}

// pageBoxes is the per-page state a commit derives from its box list: the
// lazy parent-to-children index and the motion declaration verdict for the
// display being published.
type pageBoxes struct {
	kid   kidCache
	decls displayDecls
}

// kidCache is the lazily built parent-to-children index. A commit that keeps
// the id sequence and the parent links only marks the rects stale, so the next
// kidRect rewrites them in place instead of rebuilding the map and the child
// links.
type kidCache struct {
	boxList []Box
	index   dirty.KidIndex
	// ready reports that index matches the ids and parent links of boxList.
	ready bool
	// staleRects reports that index is structurally current but its rects
	// still describe an older box list.
	staleRects bool
}

// displayDecls is the motion declaration verdict for one display. A publish
// asks for the transitions verdict while capturing and the animations verdict
// while binding, both about the same display, so one entry answers both.
type displayDecls struct {
	display     *Display
	animations  bool
	transitions bool
}

// setBoxes stores the hit-test boxes from a Redraw. It rebuilds the id index
// so lookups stay O(1) on large pages, incrementally maintains the cached
// content rect, and publishes the new list to the per-page box cache, so a
// later mark walks only the changed element's subtree. An unchanged id set
// repoints the index without clearing it, so a redraw of the same page keeps
// the map's capacity and buckets.
func (p *Page) setBoxes(boxes []Box) {
	old := p.boxes
	w, h := p.canvasSize()
	// Delta the content rect before the index is rebuilt in place: p.index
	// still describes old here, and Index clears it.
	content, ok := p.extendContent(p.index, boxes, w, h)
	sameIDs := hit.SameIDs(old, boxes)
	p.boxes = boxes
	if sameIDs {
		p.index = hit.Repoint(p.index, boxes)
	} else {
		p.index = hit.Index(p.index, boxes)
	}
	p.hits = nil
	p.finishContent(old, content, ok, w, h)
	noteBoxes(p, old, boxes, sameIDs)
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

// noteBoxes publishes a committed box list into the page's box cache. The kid
// index only records the list and whether the index built from the old one
// still applies, so a commit pays no child-index build until a mark asks for a
// subtree rect.
func noteBoxes(p *Page, old, boxes []Box, sameIDs bool) {
	boxCaches.Lock()
	e := boxCaches.byPage[p]
	if e == nil {
		e = &pageBoxes{}
		boxCaches.byPage[p] = e
	}
	e.kid.note(old, boxes, sameIDs)
	boxCaches.Unlock()
}

// note records a commit. The built index survives when the id sequence and
// the parent links are unchanged and it still describes old; only the rects
// then need the new boxes. Any other change drops the index, so the next read
// rebuilds the map and the child links.
func (k *kidCache) note(old, next []Box, sameIDs bool) {
	reusable := sameIDs && k.ready && sameBoxList(k.boxList, old) && dirty.SameParents(old, next)
	k.boxList = next
	if reusable {
		k.staleRects = true

		return
	}
	k.ready = false
	k.staleRects = false
}

// sameBoxList reports whether a and b are the same committed list: the same
// backing array with the same length. setBoxes is the only writer of the list
// it stores, so an equal head and length mean a later commit reused the list
// rather than replacing it. An empty list never matches, so its index just
// rebuilds.
func sameBoxList(a, b []Box) bool {
	if len(a) == 0 || len(a) != len(b) {
		return false
	}

	return &a[0] == &b[0]
}

// kidRect returns the union of the id's box and its descendants' boxes from
// the page's parent-to-children index, building the index on first use after a
// commit that changed the tree. A commit that only moved boxes rewrites the
// index's rects in place. ok is false when the page has no committed boxes or
// the id is absent, and the caller falls back to the id's own box.
func kidRect(p *Page, id string) (image.Rectangle, bool) {
	boxCaches.Lock()
	defer boxCaches.Unlock()

	e, found := boxCaches.byPage[p]
	if !found {
		return image.Rectangle{}, false
	}
	k := &e.kid
	switch {
	case !k.ready:
		k.index = dirty.BuildKidIndex(k.boxList)
		k.ready = true
		k.staleRects = false
	case k.staleRects:
		k.index.UpdateRects(k.boxList)
		k.staleRects = false
	}

	return k.index.Rect(id)
}

// displayDeclarations reports whether d declares any CSS animation and any CSS
// transition. One box walk answers both, and the verdict is kept while the
// same display is asked about: a publish captures transitions against the
// display it is about to install and then binds animations to it, so the pair
// costs one scan instead of two.
func (p *Page) displayDeclarations(d *Display) (animations, transitions bool) {
	if d == nil {
		return false, false
	}
	boxCaches.Lock()
	defer boxCaches.Unlock()

	e := boxCaches.byPage[p]
	if e == nil {
		e = &pageBoxes{}
		boxCaches.byPage[p] = e
	}
	if e.decls.display != d {
		a, tr := false, false
		for i := range d.Boxes {
			if !a && len(d.Boxes[i].Animations) > 0 {
				a = true
			}
			if !tr && len(d.Boxes[i].Transitions) != 0 {
				tr = true
			}
			if a && tr {
				break
			}
		}
		e.decls = displayDecls{display: d, animations: a, transitions: tr}
	}

	return e.decls.animations, e.decls.transitions
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
