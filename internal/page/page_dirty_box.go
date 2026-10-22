package page

import "image"

// labelGap is how near a label must sit to a toggle to repaint with it.
const labelGap = 12

// boxRound nudges a box edge to the nearest whole CSS pixel.
const boxRound = 0.5

// markRect queues one dirty rect on the page tracker.
func (p *Page) markRect(r image.Rectangle) {
	p.dirty.MarkRect(r)
}

// markFull dirties the whole content on the page tracker.
func (p *Page) markFull() {
	p.dirty.MarkFull()
}

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
	return image.Rect(int(b.X), int(b.Y), int(b.X+b.W+boxRound), int(b.Y+b.H+boxRound))
}

// touches reports whether two rects meet within gap pixels.
func touches(a, b image.Rectangle, gap int) bool {
	return !a.Inset(-gap).Intersect(b).Empty()
}
