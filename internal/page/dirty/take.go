package dirty

import "image"

// Take returns the regions changed since the previous Take and drains
// them. Every rect gains the Pad margin and is confined to content, then
// the list coalesces. Past a third of the content area it collapses to
// one bounding box. A whole-content change returns one content rect. ok
// is false when nothing changed.
func (d *Tracker) Take(content image.Rectangle) ([]image.Rectangle, bool) {
	if d.Pending == nil {
		d.Pending = map[string]bool{}
	}
	if d.last == nil {
		d.last = map[string]image.Rectangle{}
	}
	full := d.Full
	rects := d.Rects
	d.Full = false
	d.Rects = nil
	clear(d.Pending)

	if full {
		return append(rects[:0], content), true
	}

	rects = clampRects(rects, content)
	rects = CoalesceRects(rects, Gap)
	if len(rects) == 0 {
		return nil, false
	}

	if rectsArea(rects) > content.Dx()*content.Dy()/frameThird {
		return append(rects[:0], BoundingBox(rects)), true
	}

	return rects, true
}
