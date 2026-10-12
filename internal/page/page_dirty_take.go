package page

import "image"

// TakeDirtyRects returns the regions in CSS pixels changed since the
// previous call and clears them. Every rect gains the dirtyPad margin and
// is confined to the content rect, and the list coalesces. Past a third of
// the content area the list collapses to one bounding box. A whole-content
// change returns one content rect. ok is false when nothing changed.
func (p *Page) TakeDirtyRects() ([]image.Rectangle, bool) {
	full := p.dirtyFull
	rects := p.dirtyRects
	p.dirtyFull = false
	p.dirtyRects = nil
	clear(p.pending)

	content := p.contentRect()
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

// TakeDirty returns the region in CSS pixels changed since the previous
// call and clears it. It is the bounding box of TakeDirtyRects. A
// whole-content change returns the content rect. ok is false when nothing
// changed since the previous call.
func (p *Page) TakeDirty() (image.Rectangle, bool) {
	rects, ok := p.TakeDirtyRects()
	if !ok {
		return image.Rectangle{}, false
	}

	return boundingBox(rects), true
}
