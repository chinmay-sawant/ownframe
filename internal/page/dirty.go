package page

import (
	"github.com/chinmay-sawant/blinkless/layout"
	"github.com/chinmay-sawant/ownframe/internal/host"
	"github.com/chinmay-sawant/ownframe/internal/page/dirty"
	"github.com/chinmay-sawant/ownframe/internal/page/hit"
	"image"
	"sort"
	"sync"
)

// dirtySet is the page-facing result of one display diff. Callers in this
// package read n, rects, added, and removed. The merge lives in package dirty.
type dirtySet struct {
	rects          [dirty.MaxRects]image.Rectangle
	n              int
	added, removed int
	// memo carries the next-side op bounds the diff already computed, so
	// the perf accounting pass reuses them. It is nil when the diff took
	// a path that computed no bounds, or when the caller never asked for
	// one. Whoever consumes a non-nil memo frees it.
	memo *dirty.BoundsMemo
}

// diffDisplay returns the regions covering the operations that differ
// between prev and next. See dirty.Diff.
func diffDisplay(prev, next *Display) (dirtySet, bool, int) {
	set, full, changed := dirty.Diff(prev, next)

	return dirtySet{
		rects:   set.Rects,
		n:       set.N,
		added:   set.Added,
		removed: set.Removed,
	}, full, changed
}

// diffDisplayMemo diffs like diffDisplay and also keeps the next-side
// bounds the scan computed. See dirty.DiffWithBounds. The caller frees
// d.memo after use, on every path.
func diffDisplayMemo(prev, next *Display) (dirtySet, bool, int) {
	set, full, changed, memo := dirty.DiffWithBounds(prev, next)

	return dirtySet{
		rects:   set.Rects,
		n:       set.N,
		added:   set.Added,
		removed: set.Removed,
		memo:    memo,
	}, full, changed
}

func diffDisplayDirect(prev, next *Display) (dirtySet, bool, int) {
	set, full, changed := dirty.DiffDirect(prev, next)

	return dirtySet{
		rects:   set.Rects,
		n:       set.N,
		added:   set.Added,
		removed: set.Removed,
	}, full, changed
}

func diffDisplayDirectMemo(prev, next *Display) (dirtySet, bool, int) {
	set, full, changed, memo := dirty.DiffDirectWithBounds(prev, next)

	return dirtySet{
		rects:   set.Rects,
		n:       set.N,
		added:   set.Added,
		removed: set.Removed,
		memo:    memo,
	}, full, changed
}

// freeMemo returns the diff's bounds memo for reuse. A nil memo is a
// no-op.
func (d dirtySet) freeMemo() {
	if d.memo != nil {
		d.memo.Free()
	}
}

func opBounds(op *DisplayOp, d *Display) image.Rectangle {
	return dirty.OpBounds(op, d)
}

func sameOp(a, b *DisplayOp) bool {
	return dirty.SameOp(a, b)
}

func boxRect(b Box) image.Rectangle {
	return dirty.BoxRect(b)
}

// markRect queues one dirty rect on the page tracker and drops a cached
// picture when the rect paints something.
func markRect(p *Page, r image.Rectangle) {
	defer notifyFrame(p)
	p.tickMutated.Store(true)
	if !r.Empty() {
		p.draw.DropPNG()
	}

	p.dirty.MarkRect(r)
}

// markFull dirties the whole content on the page tracker.
func markFull(p *Page) {
	p.tickMutated.Store(true)
	p.draw.DropPNG()
	p.dirty.MarkFull()
	p.fullGen = p.generation
}

// TakeDirtyRects returns the regions in CSS pixels changed since the
// previous call and clears them. The tracker owns the damage state; the
// page only supplies its content rect. See dirty.Tracker.Take.
func (p *Page) TakeDirtyRects() ([]image.Rectangle, bool) {
	rects, ok := p.dirty.Take(contentRect(p))
	oldPainted := p.painted
	p.painted = p.draw.Display
	if oldPainted != nil && oldPainted != p.painted && oldPainted != p.draw.Display {
		releaseDisplay(p, oldPainted)
	}
	maybeReleasePainted(p)

	return rects, ok
}

// TakeDirty returns the region in CSS pixels changed since the previous
// call and clears it. It is the bounding box of TakeDirtyRects. A
// whole-content change returns the content rect. ok is false when nothing
// changed since the previous call.
func (p *Page) TakeDirty() (image.Rectangle, bool) {
	rects, ok := p.dirty.Take(contentRect(p))
	oldPainted := p.painted
	p.painted = p.draw.Display
	if oldPainted != nil && oldPainted != p.painted && oldPainted != p.draw.Display {
		releaseDisplay(p, oldPainted)
	}
	maybeReleasePainted(p)
	if !ok {
		return image.Rectangle{}, false
	}

	return dirty.BoundingBox(rects), true
}

// maybeReleasePainted drops the painted baseline after the host takes the
// damage, so a quiet page keeps one display instead of two: the current
// list. The next commit then lets the old list go instead of holding it
// alive through painted. The Take just drained every mark, so no pending
// frame can still need the baseline; a later whole-frame mark rebinds
// fullGen to its own generation before any later commit can recover.
//
// The release runs only when no later commit can need the baseline for a
// narrow recovery, and never when the baseline may still describe debt:
//   - restampRedraw or retainState: the page owes a retained-state redraw
//     or carries the retained box tree, so a re-projection or override
//     pass can advance the generation outside the commit path. Only the
//     painted display can then still recover a pending whole frame.
//   - a busy motion runtime: live players and renderers sample the
//     committed operations in place, so the painted list is still the pose
//     those samples build on.
//   - fullGen past the current generation: the outstanding whole-frame
//     mark covers an in-place tick change no display diff can see, so only
//     a full diff from the next commit stays recoverable.
//
// A stale fullGen below the current generation is fine: the Take drained
// the mark it belonged to, and any future mark refreshes the binding.
func maybeReleasePainted(p *Page) {
	if p.painted == nil {
		return
	}
	if p.restampRedraw || p.retainState {
		return
	}
	if !motionCommitIdle(p) {
		return
	}
	if p.fullGen > p.generation {
		return
	}
	if p.painted != p.draw.Display {
		old := p.painted
		p.painted = nil
		releaseDisplay(p, old)
		return
	}
	p.painted = nil
}

// Invalidate marks the region the element id changed since the previous
// call. An empty id, or an id with no box, dirties the whole frame. A
// known id takes its box, the boxes inside it, and its previous rect. A
// handler calls it for the ids it changed before the next Redraw. A page
// that never calls it stays correct: Redraw diffs the display lists.
func (p *Page) Invalidate(id string) {
	if id == "" || !hasBox(p, id) {
		markFull(p)

		return
	}

	markPending(p, id)
}

// markPending marks the id's box now and its box after the next Redraw, so
// an element that grew repaints the space it used and the space it now
// uses.
func markPending(p *Page, id string) {
	if id == "" {
		return
	}

	markID(p, id)
	p.dirty.Pend(id)
}

// markPair marks the boxes of two element ids, one before and one after a
// state change.
func markPair(p *Page, oldID, newID string) {
	markPending(p, oldID)
	markPending(p, newID)
}

// markID marks the box an id names, or the whole frame when no box has it.
func markID(p *Page, id string) {
	if id == "" {
		return
	}

	if !hasBox(p, id) {
		markFull(p)

		return
	}

	markElement(p, id)
}

// markElement marks the id's box, the boxes inside its element subtree,
// and the previous rect for the same id, so an element that moved
// repaints the space it left. The subtree walk replaces the legacy full
// box scan: a box geometrically inside the id's rect adds nothing to
// the union, so only descendants outside it change the region, and
// those repaint with the parent rather than being missed.
func markElement(p *Page, id string) {
	r := subtreeRect(p, id)
	markRect(p, r)
	markRect(p, p.dirty.Prev(id))
	p.dirty.Note(id, r)
}

// subtreeRect returns the union of the id's box and its descendants'
// boxes from the parent-to-children index, falling back to the id's own
// box when the index has no entry for it.
func subtreeRect(p *Page, id string) image.Rectangle {
	if r, ok := kidRect(p, id); ok {
		return r
	}
	return dirty.BoxRect(boxByID(p, id))
}

// markToggle marks a checkbox or radio, its label box when the layout has
// one, and a halo that covers an inline label it does not.
func markToggle(p *Page, id string) {
	markID(p, id)

	ctl := dirty.BoxRect(boxByID(p, id))
	markRect(p, dirty.ToggleHalo(ctl))

	for _, b := range p.boxes {
		if r, ok := dirty.LabelNear(ctl, b); ok {
			markRect(p, r)
		}
	}
}

// dirtyFromDisplay adds the difference between prev and display to the
// dirty region. The first draw and a size change dirty the frame. With Perf
// on it also stores the changed, dirty, and region counters Stats reports;
// TakeDirty clears only the region, never these counters.
func dirtyFromDisplay(p *Page, prev, display *Display) {
	dirtyFromDisplayWith(p, prev, display, diffDisplay, diffDisplayMemo)
}

func dirtyFromPatchedDisplay(p *Page, prev, display *Display) {
	dirtyFromDisplayWith(p, prev, display, diffDisplayDirect, diffDisplayDirectMemo)
}

func dirtyFromDisplayWith(
	p *Page,
	prev, display *Display,
	diffFn, diffMemoFn func(*Display, *Display) (dirtySet, bool, int),
) {
	if prev == nil {
		markFull(p)
		if p.perf {
			recordFull(p, len(display.Ops))
		}
		p.lastPatch = DisplayPatch{
			Changed: len(display.Ops),
			Added:   len(display.Ops),
			Regions: 1,
			Full:    true,
		}

		return
	}

	// A full mark left from an earlier display must not hide this update's
	// own change set. When a display was painted since that mark, diffing
	// against it recovers every change the host has not seen and can often
	// narrow the frame to the region the update actually touched. Without a
	// painted display there is no baseline to recover from, so the frame
	// stays full; the report still describes this update instead of the one
	// that first marked full.
	if p.dirty.Full {
		if p.painted != nil && p.fullGen < p.generation {
			d, full, changed := diffMemoFn(p.painted, display)
			if !full {
				p.dirty.Full = false
				markDisplayPartial(p, d, display, changed)

				return
			}
			// The recovery diff is still full, so the frame stays full.
			// Its counts already describe everything the host has not
			// seen, so no second diff runs: one diff carries the area
			// budget that decides full, and its result is the report.
			d.freeMemo()
			markFull(p)
			// This commit carries a mark that was already pending, so a later
			// diff from painted cannot recover what the mark covered (an
			// in-place tick, for example). Bind the mark to the new generation;
			// only a full diff from this commit stays recoverable.
			p.fullGen = p.generation + 1
			if p.perf {
				recordFullChanged(p, changed, len(display.Ops))
			}
			p.lastPatch = DisplayPatch{
				Changed: changed,
				Added:   d.added,
				Removed: d.removed,
				Regions: 1,
				Full:    true,
			}

			return
		}

		d, _, changed := diffFn(prev, display)
		markFull(p)
		// This commit carries a mark that was already pending, so a later
		// diff from painted cannot recover what the mark covered (an
		// in-place tick, for example). Bind the mark to the new generation;
		// only a full diff from this commit stays recoverable.
		p.fullGen = p.generation + 1
		if p.perf {
			recordFullChanged(p, changed, len(display.Ops))
		}
		p.lastPatch = DisplayPatch{
			Changed: changed,
			Added:   d.added,
			Removed: d.removed,
			Regions: 1,
			Full:    true,
		}

		return
	}

	d, full, changed := diffMemoFn(prev, display)
	if full {
		d.freeMemo()
		markFull(p)
		if p.perf {
			recordFullChanged(p, changed, len(display.Ops))
		}
		p.lastPatch = DisplayPatch{
			Changed: changed,
			Added:   d.added,
			Removed: d.removed,
			Regions: 1,
			Full:    true,
		}

		return
	}

	markDisplayPartial(p, d, display, changed)
}

// markDisplayPartial queues the regions one partial display diff found and
// records the counters and report that describe it. It frees the diff's
// bounds memo on every path.
func markDisplayPartial(p *Page, d dirtySet, display *Display, changed int) {
	defer d.freeMemo()
	for i := range d.n {
		markRect(p, d.rects[i])
	}

	if p.perf {
		if d.memo != nil && d.memo.Len() == len(display.Ops) {
			recordPartialBounds(p, changed, d.rects[:d.n], d.memo, display)
		} else {
			recordPartial(p, changed, d.rects[:d.n], display)
		}
	}

	p.lastPatch = DisplayPatch{
		Changed: changed,
		Added:   d.added,
		Removed: d.removed,
		Regions: d.n,
		Full:    false,
	}
}

// applyPending marks the new box of every pending id, so an element that
// grew or shrank repaints the space it now uses too, then clears the map
// for reuse instead of dropping it.
func applyPending(p *Page) {
	for id := range p.dirty.Pending {
		markElement(p, id)
	}

	clear(p.dirty.Pending)
}

// recordFull stores whole-frame dirty counters for a first draw, a size
// change, or a paint fallback. n is the ops or boxes now on screen.
func recordFull(p *Page, n int) {
	recordFullChanged(p, n, n)
}

// recordFullChanged is recordFull with the operations that actually differ
// kept separate from the operations the full repaint meets. A full damage
// region still counts every operation as dirty; only the changed count is
// the display diff's own result.
func recordFullChanged(p *Page, changed, n int) {
	p.stats.changedOps = changed
	p.stats.dirtyOps = n
	p.stats.dirtyRegions = 1
	r := contentRect(p)
	p.stats.dirtyArea = r.Dx() * r.Dy()
}

// recordPartial counts the diff regions and each intersecting operation once.
func recordPartial(p *Page, changed int, rects []image.Rectangle, d *Display) {
	p.stats.changedOps = changed
	p.stats.dirtyOps = 0
	p.stats.dirtyRegions = len(rects)
	p.stats.dirtyArea = 0
	if len(rects) == 0 {
		return
	}
	for _, r := range rects {
		p.stats.dirtyArea += r.Dx() * r.Dy()
	}
	for i := range d.Ops {
		bounds := dirty.OpBounds(&d.Ops[i], d)
		for _, r := range rects {
			if bounds.Overlaps(r) {
				p.stats.dirtyOps++
				break
			}
		}
	}
}

// recordPartialBounds counts like recordPartial, but reuses the bounds
// the display diff already computed instead of running a second bounds
// pass over every operation. An index the diff never computed falls
// back to OpBounds, so the counts always equal recordPartial's.
func recordPartialBounds(p *Page, changed int, rects []image.Rectangle, memo *dirty.BoundsMemo, d *Display) {
	p.stats.changedOps = changed
	p.stats.dirtyOps = 0
	p.stats.dirtyRegions = len(rects)
	p.stats.dirtyArea = 0
	if len(rects) == 0 {
		return
	}
	for _, r := range rects {
		p.stats.dirtyArea += r.Dx() * r.Dy()
	}
	for i := range d.Ops {
		bounds, ok := memo.Get(i)
		if !ok {
			bounds = dirty.OpBounds(&d.Ops[i], d)
		}
		for _, r := range rects {
			if bounds.Overlaps(r) {
				p.stats.dirtyOps++
				break
			}
		}
	}
}

// boxCache is the page-owned box cache: the lazy kid index and the motion
// declaration verdict derived from the committed box list. It is allocated on
// first use and dropped by forgetBoxes, so a discarded page takes its cache
// with it to the collector. mu keeps the old process-wide cache's guarantee
// that noteBoxes, kidRect, displayDeclarations, and forgetBoxes never race.
type boxCache struct {
	mu      sync.Mutex
	entries *pageBoxes
}

// lockedBoxes returns the page's cache entry with the cache mutex held,
// allocating it when create is set. The caller must call p.boxCache.mu.Unlock.
// A nil entry (create false, nothing committed) answers like a page with no
// committed boxes.
func lockedBoxes(p *Page, create bool) *pageBoxes {
	p.boxCache.mu.Lock()
	if p.boxCache.entries == nil && create {
		p.boxCache.entries = &pageBoxes{}
	}

	return p.boxCache.entries
}

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
func setBoxes(p *Page, boxes []Box) {
	old := p.boxes
	w, h := canvasSize(p)
	sameBacking := sameBoxList(old, boxes)
	if sameBacking && p.contentOK && p.contentW == w && p.contentH == h {
		p.hits = nil
		return
	}
	sameIDs := sameBacking || hit.SameIDs(old, boxes)
	var content image.Rectangle
	var ok bool
	if sameIDs && !sameBacking {
		content, ok = extendContentSameIDs(p, old, boxes, w, h)
	} else {
		// Delta the content rect before the index is rebuilt in place: p.index
		// still describes old here, and Index clears it.
		content, ok = extendContent(p, p.index, boxes, w, h)
	}
	p.boxes = boxes
	switch {
	case sameBacking:
		// Same backing array and ID order: p.index already points to &boxes[i].
	case sameIDs:
		p.index = hit.Repoint(p.index, boxes)
	default:
		p.index = hit.Index(p.index, boxes)
	}
	p.hits = nil
	finishContent(p, old, content, ok, sameIDs, w, h)
	noteBoxes(p, old, boxes, sameIDs)
}

// canvasSize is the canvas the content rect covers: the committed
// display's size, or the window size before the first commit.
func canvasSize(p *Page) (int, int) {
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
// extendContentSameIDs updates the cached content rect by comparing old and
// boxes at matching indices without map lookups. Callers must ensure old and
// boxes have the same length, matching IDs, and distinct backing arrays.
func extendContentSameIDs(p *Page, old, boxes []Box, w, h int) (image.Rectangle, bool) {
	if !p.contentOK || p.contentW != w || p.contentH != h {
		return image.Rectangle{}, false
	}
	content := p.content
	for i := range boxes {
		o, b := &old[i], &boxes[i]
		if o.X == b.X && o.Y == b.Y && o.W == b.W && o.H == b.H {
			continue
		}
		or, r := boxRect(*o), boxRect(*b)
		if r == or {
			continue
		}
		if !or.In(r) {
			// If the old rect was strictly inside the cached content rect,
			// shrinking or moving it cannot shrink the union's outer edges.
			if or.Min.X <= p.content.Min.X || or.Min.Y <= p.content.Min.Y ||
				or.Max.X >= p.content.Max.X || or.Max.Y >= p.content.Max.Y {
				return content, false
			}
		}
		content = content.Union(r)
	}
	return content, true
}

func extendContent(p *Page, oldIndex map[string]*Box, boxes []Box, w, h int) (image.Rectangle, bool) {
	if !p.contentOK || oldIndex == nil || p.contentW != w || p.contentH != h {
		return image.Rectangle{}, false
	}
	content := p.content
	for i := range boxes {
		b := &boxes[i]
		r := boxRect(*b)
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
func finishContent(p *Page, old []Box, content image.Rectangle, ok, sameIDs bool, w, h int) {
	if ok && !sameIDs {
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
	for i := range boxes {
		r = r.Union(boxRect(boxes[i]))
	}
	return r
}

// noteBoxes publishes a committed box list into the page's box cache. The kid
// index only records the list and whether the index built from the old one
// still applies, so a commit pays no child-index build until a mark asks for a
// subtree rect.
func noteBoxes(p *Page, old, boxes []Box, sameIDs bool) {
	e := lockedBoxes(p, true)
	e.kid.note(old, boxes, sameIDs)
	p.boxCache.mu.Unlock()
}

// forgetBoxes drops the page's box-cache entry, if any. It is a no-op on a
// nil page or a page with no entry. invalidateCache calls it when the parsed
// tree is discarded, so a source change does not keep the old box list alive
// until the next commit. A missing entry answers like a page with no
// committed boxes: kidRect misses and the caller falls back, and
// displayDeclarations and the next noteBoxes rebuild the entry.
func forgetBoxes(p *Page) {
	if p == nil {
		return
	}
	p.boxCache.mu.Lock()
	p.boxCache.entries = nil
	p.boxCache.mu.Unlock()
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
	e := lockedBoxes(p, false)
	defer p.boxCache.mu.Unlock()
	if e == nil {
		return image.Rectangle{}, false
	}
	k := &e.kid
	switch {
	case !k.ready:
		k.index.Rebuild(k.boxList)
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
// carryDisplayDeclarations transfers the cached animation/transition verdict
// from prev to next when a paint-only splice reused prev.Boxes in place.
func carryDisplayDeclarations(p *Page, prev, next *Display) {
	if prev == nil || next == nil {
		return
	}
	e := lockedBoxes(p, false)
	defer p.boxCache.mu.Unlock()
	if e != nil && e.decls.display == prev {
		e.decls.display = next
	}
}

func displayDeclarations(p *Page, d *Display) (animations, transitions bool) {
	if d == nil {
		return false, false
	}
	e := lockedBoxes(p, true)
	defer p.boxCache.mu.Unlock()
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

func setPlacedBoxes(p *Page, d *Display) {
	fixed := fixedBoxes(d)
	if len(fixed) == 0 {
		setBoxes(p, d.Boxes)
		p.hits = nil

		return
	}
	if hitsReusable(p, d.Boxes, fixed) {
		// Same elements, same rectangles, same pins: the stored list already
		// describes the new display, so keep it and its capacity instead of
		// recopying every item. The index still repoints at the new boxes.
		keep := p.hits
		setBoxes(p, d.Boxes)
		p.hits = keep

		return
	}
	hits := p.hits
	setBoxes(p, d.Boxes)
	p.hits = hit.List(hits, d.Boxes, fixed)
}

// hitsReusable reports whether the stored hit list already describes boxes
// with the same pins: the same elements with the same rectangles in order,
// and the same fixed roots as sets. The markings are a pure function of the
// box links and the roots, so equal inputs mean an identical list.
func hitsReusable(p *Page, boxes, fixed []Box) bool {
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

// BoxByID returns the first placed box matching id without cloning the full
// boxes slice.
func (p *Page) BoxByID(id string) (Box, bool) {
	if p == nil || id == "" {
		return Box{}, false
	}
	for i := range p.boxes {
		if p.boxes[i].ID == id {
			return p.boxes[i], true
		}
	}
	return Box{}, false
}

// BoxesSlice returns the last hit-test boxes directly without copying, or nil
// when nothing has been drawn. The returned slice is read-only.
func (p *Page) BoxesSlice() []Box {
	if p == nil {
		return nil
	}

	return p.boxes
}

// BoxAt returns the innermost box containing x, y, or (Box{}, false) when no
// box contains the point.
func (p *Page) BoxAt(x, y float64) (Box, bool) {
	if p == nil {
		return Box{}, false
	}

	if p.hasViewportSticky {
		return hitBox(p, x, y)
	}
	return hit.Hit(p.boxes, x, y)
}

var _ host.BoxSlicer = (*Page)(nil)

// boxByID returns the last box laid out with that id, or Box{ID: id}. The
// lookup is one map read, never a walk over p.boxes.
func boxByID(p *Page, id string) Box {
	return hit.ByID(p.index, id)
}

// hasBox reports whether the last layout has a box for id, in one map read.
func hasBox(p *Page, id string) bool {
	return hit.Has(p.index, id)
}

// contentRect is the painted page in CSS pixels: the canvas plus any box
// that overflows it, the region the window's replay buffer covers. The
// rect is cached per canvas size until the next setBoxes, so the
// per-frame TakeDirty costs O(1) on large pages. setBoxes maintains it
// incrementally; the size check stays as the backstop for restores that
// bypass setBoxes.
func contentRect(p *Page) image.Rectangle {
	w, h := canvasSize(p)
	if p.contentOK && p.contentW == w && p.contentH == h {
		return p.content
	}
	r := image.Rect(0, 0, w, h).Union(unionBoxes(p.boxes))
	p.content, p.contentW, p.contentH, p.contentOK = r, w, h, true

	return r
}

var _ host.LayerSource = (*Page)(nil)

// LayerSpecs reports this frame's promoted compositor layers. It is the page
// side of host.LayerSource: the window asks once per tile frame, before it
// plans repaints, so the renderer can bake a promoted element once and
// composite it through the animation or transition matrix instead of replaying
// its operations every frame.
//
// display must be the list the page is drawing. A mismatch means the window
// holds a list the page has replaced, so the box indices the specifications
// name no longer describe the operations the window would paint; the direct
// replay path keeps the frame.
//
// The motion runtime merges the animation and transition registries before this
// method sees them: a live transform or opacity transition becomes one entry,
// and a transition and an animation on the same element share one entry, hence
// one layer. The runtime already drops inactive, paused, layout-moving,
// restamping, 3D, and projection transitions. This method adds the rule a layer
// specification needs that the runtime cannot make from a display: the matrix
// must be invertible so the layer texture can hold the captured pose.
//
// An entry that also writes opacity keeps its Alpha, which is the transition's
// group opacity when the transition owns the field. The renderer divides that
// sampled group opacity out of the bake and applies it once at composite time,
// so a transform and an opacity transition can promote together and the
// overlapping children of an opacity transition fade once, not once per
// operation. A pure opacity entry has an identity matrix and promotes on Alpha
// alone.
//
// Release is not this method's job: a transition that ended or was cancelled
// stops being listed, and the renderer's BeginLayers releases the layer for the
// missing specification and damages its last pose. A page with the animation
// driver off still promotes transitions, because the two switches are
// independent.
//
// The runtime reuses its entry slice, so every field is copied into a new slice
// here. The window consumes the result before the next frame asks again.
// SetPromotedLayers tells the motion driver which elements the compositor
// kept as retained layers on the last frame. Their paint changes land in the
// layer textures, so the next Step leaves their page damage to the layer
// cache; a demotion reports the pose it leaves behind.
func (p *Page) SetPromotedLayers(ids []string) {
	if p == nil || p.motion == nil {
		return
	}

	p.motion.SetPromotedLayers(ids)
}

// SetPromotedLayerBoxes tells the motion driver which box indices (including
// anonymous elements without an id) the compositor kept as retained layers.
func (p *Page) SetPromotedLayerBoxes(boxes []int) {
	if p == nil || p.motion == nil {
		return
	}

	p.motion.SetPromotedLayerBoxes(boxes)
}

func (p *Page) LayerSpecs(display *layout.Display) []host.LayerSpec {
	if p == nil || p.motion == nil || (p.animationsOff && p.transitionsOff) {
		return nil
	}
	if display == nil || display != p.draw.Display {
		return nil
	}

	entries := p.motion.LayerEntriesFor(display)
	if len(entries) == 0 {
		return nil
	}

	specs := make([]host.LayerSpec, 0, len(entries))
	for i := range entries {
		entry := &entries[i]
		base, ok := host.InvertLayerMatrix(entry.Matrix)
		if !ok {
			continue
		}
		specs = append(specs, host.LayerSpec{
			ElementID: entry.ElementID,
			BoxIndex:  entry.BoxIndex,
			Matrix:    entry.Matrix,
			Base:      base,
			Alpha:     entry.Alpha,
			Mode:      string(entry.Mode),
		})
	}
	if len(specs) == 0 {
		return nil
	}

	return specs
}

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
	refreshStickyBoxes(p, x, y, width, height)
}

func refreshStickyBoxes(p *Page, x, y, width, height float64) {
	display := p.Display()
	if !p.hasViewportSticky || display == nil {
		return
	}
	setStickyBoxes(p, layout.StickyView(display, x, y, width, height))
}

func setDisplayBoxes(p *Page, display *Display) {
	p.hasViewportSticky = layout.HasViewportSticky(display)
	if !p.hasViewportSticky {
		p.stickyViewport = nil
		setPlacedBoxes(p, display)
		return
	}
	viewport := scrollViewport{x: float64(p.scroll.Windowing.X), y: float64(p.scroll.Windowing.Y), width: float64(p.width), height: float64(p.height)}
	if p.stickyViewport != nil {
		viewport = *p.stickyViewport
	}
	setStickyBoxes(p, layout.StickyView(display, viewport.x, viewport.y, viewport.width, viewport.height))
}

// A sticky stacking context can paint above a later in-flow sibling. Preserve
// document order in Boxes, but route pointer input in the engine's paint order.
func setStickyBoxes(p *Page, display *Display) {
	fixed := fixedBoxes(display)
	setBoxes(p, display.Boxes)
	items := hit.List(nil, display.Boxes, fixed)
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
