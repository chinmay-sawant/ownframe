package page

import (
	"image"

	"github.com/chinmay-sawant/ownframe/internal/page/dirty"
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
func (p *Page) markRect(r image.Rectangle) {
	defer p.notifyFrame()
	p.tickMutated.Store(true)
	if !r.Empty() {
		p.draw.DropPNG()
	}

	p.dirty.MarkRect(r)
}

// markFull dirties the whole content on the page tracker.
func (p *Page) markFull() {
	p.tickMutated.Store(true)
	p.draw.DropPNG()
	p.dirty.MarkFull()
	p.fullGen = p.generation
}

// TakeDirtyRects returns the regions in CSS pixels changed since the
// previous call and clears them. The tracker owns the damage state; the
// page only supplies its content rect. See dirty.Tracker.Take.
func (p *Page) TakeDirtyRects() ([]image.Rectangle, bool) {
	rects, ok := p.dirty.Take(p.contentRect())
	p.painted = p.draw.Display
	p.maybeReleasePainted()

	return rects, ok
}

// TakeDirty returns the region in CSS pixels changed since the previous
// call and clears it. It is the bounding box of TakeDirtyRects. A
// whole-content change returns the content rect. ok is false when nothing
// changed since the previous call.
func (p *Page) TakeDirty() (image.Rectangle, bool) {
	rects, ok := p.dirty.Take(p.contentRect())
	p.painted = p.draw.Display
	p.maybeReleasePainted()
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
func (p *Page) maybeReleasePainted() {
	if p.painted == nil {
		return
	}
	if p.restampRedraw || p.retainState {
		return
	}
	if !p.motionCommitIdle() {
		return
	}
	if p.fullGen > p.generation {
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
	if id == "" || !p.hasBox(id) {
		p.markFull()

		return
	}

	p.markPending(id)
}

// markPending marks the id's box now and its box after the next Redraw, so
// an element that grew repaints the space it used and the space it now
// uses.
func (p *Page) markPending(id string) {
	if id == "" {
		return
	}

	p.markID(id)
	p.dirty.Pend(id)
}

// markPair marks the boxes of two element ids, one before and one after a
// state change.
func (p *Page) markPair(oldID, newID string) {
	p.markPending(oldID)
	p.markPending(newID)
}

// markID marks the box an id names, or the whole frame when no box has it.
func (p *Page) markID(id string) {
	if id == "" {
		return
	}

	if !p.hasBox(id) {
		p.markFull()

		return
	}

	p.markElement(id)
}

// markElement marks the id's box, the boxes inside its element subtree,
// and the previous rect for the same id, so an element that moved
// repaints the space it left. The subtree walk replaces the legacy full
// box scan: a box geometrically inside the id's rect adds nothing to
// the union, so only descendants outside it change the region, and
// those repaint with the parent rather than being missed.
func (p *Page) markElement(id string) {
	r := p.subtreeRect(id)
	p.markRect(r)
	p.markRect(p.dirty.Prev(id))
	p.dirty.Note(id, r)
}

// subtreeRect returns the union of the id's box and its descendants'
// boxes from the parent-to-children index, falling back to the id's own
// box when the index has no entry for it.
func (p *Page) subtreeRect(id string) image.Rectangle {
	if r, ok := kidRect(p, id); ok {
		return r
	}
	return dirty.BoxRect(p.boxByID(id))
}

// markToggle marks a checkbox or radio, its label box when the layout has
// one, and a halo that covers an inline label it does not.
func (p *Page) markToggle(id string) {
	p.markID(id)

	ctl := dirty.BoxRect(p.boxByID(id))
	p.markRect(dirty.ToggleHalo(ctl))

	for _, b := range p.boxes {
		if r, ok := dirty.LabelNear(ctl, b); ok {
			p.markRect(r)
		}
	}
}

// dirtyFromDisplay adds the difference between prev and display to the
// dirty region. The first draw and a size change dirty the frame. With Perf
// on it also stores the changed, dirty, and region counters Stats reports;
// TakeDirty clears only the region, never these counters.
func (p *Page) dirtyFromDisplay(prev, display *Display) {
	if prev == nil {
		p.markFull()
		if p.perf {
			p.recordFull(len(display.Ops))
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
			d, full, changed := diffDisplayMemo(p.painted, display)
			if !full {
				p.dirty.Full = false
				p.markDisplayPartial(d, display, changed)

				return
			}
			// The recovery diff is still full, so the frame stays full.
			// Its counts already describe everything the host has not
			// seen, so no second diff runs: one diff carries the area
			// budget that decides full, and its result is the report.
			d.freeMemo()
			p.markFull()
			// This commit carries a mark that was already pending, so a later
			// diff from painted cannot recover what the mark covered (an
			// in-place tick, for example). Bind the mark to the new generation;
			// only a full diff from this commit stays recoverable.
			p.fullGen = p.generation + 1
			if p.perf {
				p.recordFullChanged(changed, len(display.Ops))
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

		d, _, changed := diffDisplay(prev, display)
		p.markFull()
		// This commit carries a mark that was already pending, so a later
		// diff from painted cannot recover what the mark covered (an
		// in-place tick, for example). Bind the mark to the new generation;
		// only a full diff from this commit stays recoverable.
		p.fullGen = p.generation + 1
		if p.perf {
			p.recordFullChanged(changed, len(display.Ops))
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

	d, full, changed := diffDisplayMemo(prev, display)
	if full {
		d.freeMemo()
		p.markFull()
		if p.perf {
			p.recordFullChanged(changed, len(display.Ops))
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

	p.markDisplayPartial(d, display, changed)
}

// markDisplayPartial queues the regions one partial display diff found and
// records the counters and report that describe it. It frees the diff's
// bounds memo on every path.
func (p *Page) markDisplayPartial(d dirtySet, display *Display, changed int) {
	defer d.freeMemo()
	for i := range d.n {
		p.markRect(d.rects[i])
	}

	if p.perf {
		if d.memo != nil && d.memo.Len() == len(display.Ops) {
			p.recordPartialBounds(changed, d.rects[:d.n], d.memo, display)
		} else {
			p.recordPartial(changed, d.rects[:d.n], display)
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
func (p *Page) applyPending() {
	for id := range p.dirty.Pending {
		p.markElement(id)
	}

	clear(p.dirty.Pending)
}

// recordFull stores whole-frame dirty counters for a first draw, a size
// change, or a paint fallback. n is the ops or boxes now on screen.
func (p *Page) recordFull(n int) {
	p.recordFullChanged(n, n)
}

// recordFullChanged is recordFull with the operations that actually differ
// kept separate from the operations the full repaint meets. A full damage
// region still counts every operation as dirty; only the changed count is
// the display diff's own result.
func (p *Page) recordFullChanged(changed, n int) {
	p.stats.changedOps = changed
	p.stats.dirtyOps = n
	p.stats.dirtyRegions = 1
	r := p.contentRect()
	p.stats.dirtyArea = r.Dx() * r.Dy()
}

// recordPartial counts the diff regions and each intersecting operation once.
func (p *Page) recordPartial(changed int, rects []image.Rectangle, d *Display) {
	p.stats.changedOps = changed
	p.stats.dirtyOps = 0
	p.stats.dirtyRegions = len(rects)
	p.stats.dirtyArea = 0
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
func (p *Page) recordPartialBounds(changed int, rects []image.Rectangle, memo *dirty.BoundsMemo, d *Display) {
	p.stats.changedOps = changed
	p.stats.dirtyOps = 0
	p.stats.dirtyRegions = len(rects)
	p.stats.dirtyArea = 0
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
