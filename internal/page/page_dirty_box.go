package page

import (
	"github.com/chinmay-sawant/ownframe/internal/page/dirty"
	"image"
)

// labelGap is how near a label must sit to a toggle to repaint with it.
const labelGap = 12

// boxRound nudges a box edge to the nearest whole CSS pixel.
const boxRound = 0.5

// markRect queues one dirty rect on the page tracker.
func (p *Page) markRect(r image.Rectangle) {
	if !r.Empty() {
		p.draw.DropPNG()
	}
	p.dirty.MarkRect(r)
}

// markFull dirties the whole content on the page tracker.
func (p *Page) markFull() {
	p.draw.DropPNG()
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

// TakeDirtyRects returns the regions in CSS pixels changed since the
// previous call and clears them. The tracker owns the damage state; the
// page only supplies its content rect. See dirty.Tracker.Take.
func (p *Page) TakeDirtyRects() ([]image.Rectangle, bool) {
	return p.dirty.Take(p.contentRect())
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

	return dirty.BoundingBox(rects), true
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

// markElement marks the id's box, the boxes inside it, and the previous
// rect for the same id, so an element that moved repaints the space it left.
func (p *Page) markElement(id string) {
	r := p.growBox(boxRect(p.boxByID(id)))
	p.markRect(r)
	p.markRect(p.dirty.Prev(id))
	p.dirty.Note(id, r)
}

// markToggle marks a checkbox or radio, its label box when the layout has
// one, and a halo that covers an inline label it does not.
func (p *Page) markToggle(id string) {
	p.markID(id)

	ctl := boxRect(p.boxByID(id))
	p.markRect(ctl.Inset(-labelGap))

	for _, b := range p.boxes {
		if b.Tag == "label" && touches(ctl, boxRect(b), labelGap) {
			p.markRect(boxRect(b))
		}
	}
}
