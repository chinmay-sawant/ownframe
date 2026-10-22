package page

import (
	"image"

	"github.com/chinmay-sawant/ownframe/internal/page/dirty"
)

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
