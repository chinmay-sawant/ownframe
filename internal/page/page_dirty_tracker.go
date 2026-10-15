package page

import "image"

// dirtyTracker owns the damage state a Page delegates to: the whole-frame
// flag, the queued rects, the pending element ids, and the last rect per
// id. It needs no page fixture, so tests drive it directly.
type dirtyTracker struct {
	pending map[string]bool
	last    map[string]image.Rectangle
	rects   []image.Rectangle
	full    bool
}

// markRect queues one rect, merging neighbours within dirtyGap. Past
// maxDirtyRects the list collapses to its bounding box.
func (d *dirtyTracker) markRect(r image.Rectangle) {
	if r.Empty() {
		return
	}

	d.rects = append(d.rects, r)
	d.rects = coalesceRects(d.rects, dirtyGap)
	if len(d.rects) > maxDirtyRects {
		d.rects[0] = boundingBox(d.rects)
		d.rects = d.rects[:1]
	}
}

// markFull dirties the whole content; a later take returns it.
func (d *dirtyTracker) markFull() {
	d.full = true
}

// pend remembers id for the next applyPending, keeping the map for reuse.
func (d *dirtyTracker) pend(id string) {
	if d.pending == nil {
		d.pending = map[string]bool{}
	}

	d.pending[id] = true
}

// prev returns the last rect marked for id, or zero when never marked.
func (d *dirtyTracker) prev(id string) image.Rectangle {
	return d.last[id]
}

// note records the last rect marked for id.
func (d *dirtyTracker) note(id string, r image.Rectangle) {
	if d.last == nil {
		d.last = map[string]image.Rectangle{}
	}

	d.last[id] = r
}
