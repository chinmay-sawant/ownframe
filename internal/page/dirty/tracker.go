package dirty

import "image"

// Tracker owns the damage state a Page delegates to: the whole-frame
// flag, the queued rects, the pending element ids, and the last rect per
// id. It needs no page fixture, so tests drive it directly.
type Tracker struct {
	Pending map[string]bool
	last    map[string]image.Rectangle
	Rects   []image.Rectangle
	Full    bool
}

// MarkRect queues one rect, merging neighbours within Gap. Past
// MaxRects the list collapses to its bounding box.
func (d *Tracker) MarkRect(r image.Rectangle) {
	if r.Empty() {
		return
	}

	d.Rects = append(d.Rects, r)
	d.Rects = CoalesceRects(d.Rects, Gap)
	if len(d.Rects) > MaxRects {
		d.Rects[0] = BoundingBox(d.Rects)
		d.Rects = d.Rects[:1]
	}
}

// MarkFull dirties the whole content; a later Take returns it.
func (d *Tracker) MarkFull() {
	d.Full = true
}

// Pend remembers id for the next applyPending, keeping the map for reuse.
func (d *Tracker) Pend(id string) {
	if d.Pending == nil {
		d.Pending = map[string]bool{}
	}

	d.Pending[id] = true
}

// Prev returns the last rect marked for id, or zero when never marked.
func (d *Tracker) Prev(id string) image.Rectangle {
	return d.last[id]
}

// Note records the last rect marked for id.
func (d *Tracker) Note(id string, r image.Rectangle) {
	if d.last == nil {
		d.last = map[string]image.Rectangle{}
	}

	d.last[id] = r
}
