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
	d.Full = false
	clear(d.Pending)

	d.takeBuf = append(d.takeBuf[:0], d.Rects...)
	d.Rects = d.Rects[:0]
	rects := d.takeBuf

	if full {
		rects = append(rects[:0], content)
		d.takeBuf = rects
		return d.takeBuf, true
	}

	rects = clampRects(rects, content)
	rects = CoalesceRects(rects, Gap)
	if len(rects) == 0 {
		return nil, false
	}

	if rectsArea(rects) > content.Dx()*content.Dy()/frameThird {
		box := BoundingBox(rects)
		rects = append(rects[:0], box)
		d.takeBuf = rects
		return d.takeBuf, true
	}
	d.takeBuf = rects
	return d.takeBuf, true
}
