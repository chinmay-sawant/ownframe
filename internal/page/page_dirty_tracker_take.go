package page

import "image"

// take returns the regions changed since the previous take and drains them.
// Every rect gains the dirtyPad margin and is confined to content, then the
// list coalesces. Past a third of the content area it collapses to one
// bounding box. A whole-content change returns one content rect. ok is
// false when nothing changed.
func (d *dirtyTracker) take(content image.Rectangle) ([]image.Rectangle, bool) {
	if d.pending == nil {
		d.pending = map[string]bool{}
	}
	if d.last == nil {
		d.last = map[string]image.Rectangle{}
	}
	full := d.full
	rects := d.rects
	d.full = false
	d.rects = nil
	clear(d.pending)

	if full {
		return append(rects[:0], content), true
	}

	rects = clampRects(rects, content)
	rects = coalesceRects(rects, dirtyGap)
	if len(rects) == 0 {
		return nil, false
	}

	if rectsArea(rects) > content.Dx()*content.Dy()/3 {
		return append(rects[:0], boundingBox(rects)), true
	}

	return rects, true
}
