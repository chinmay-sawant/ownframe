package page

import "image"

// dirtyPad is the margin a dirty rect gains for shadows, outlines, and
// antialiasing.
const dirtyPad = 4

// dirtyGap is how near two dirty rects must sit to merge into one.
const dirtyGap = 8

// labelGap is how near a label must sit to a toggle to repaint with it.
const labelGap = 12

// growBox covers the boxes inside r, so a nested span repaints with its
// parent.
func (p *Page) growBox(r image.Rectangle) image.Rectangle {
	if r.Empty() {
		return r
	}

	for _, b := range p.boxes {
		inner := boxRect(b)
		if !inner.Empty() && inner.In(r) {
			r = r.Union(inner)
		}
	}

	return r
}

// boxRect is one box in whole CSS pixels.
func boxRect(b Box) image.Rectangle {
	return image.Rect(int(b.X), int(b.Y), int(b.X+b.W+0.5), int(b.Y+b.H+0.5))
}

// touches reports whether two rects meet within gap pixels.
func touches(a, b image.Rectangle, gap int) bool {
	return !a.Inset(-gap).Intersect(b).Empty()
}

// markRect inserts one dirty rect, merging it with every existing rect it
// meets within dirtyGap. Past maxDirtyRects the list collapses to its
// bounding box, which trades a bigger repaint for a shorter list.
func (p *Page) markRect(r image.Rectangle) {
	if r.Empty() {
		return
	}

	p.dirtyRects = append(p.dirtyRects, r)
	p.dirtyRects = coalesceRects(p.dirtyRects, dirtyGap)
	if len(p.dirtyRects) > maxDirtyRects {
		p.dirtyRects[0] = boundingBox(p.dirtyRects)
		p.dirtyRects = p.dirtyRects[:1]
	}
}

// markFull dirties the whole content; a later TakeDirty returns the
// content rect.
func (p *Page) markFull() {
	p.dirtyFull = true
}
