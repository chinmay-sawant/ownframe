package motion

import (
	"image"

	"github.com/chinmay-sawant/ownframe/internal/page/dirty"
)

// Damage is the set of rectangles one Step changed, mirroring the page's
// dirty set. It keeps at most MaxRects of them: past the cap it collapses to
// one bounding box and Overflow reports that. A caller reads Rects[:N].
type Damage struct {
	Rects    [dirty.MaxRects]image.Rectangle
	N        int
	Overflow bool

	union image.Rectangle
}

// Add inserts one changed rectangle, merging it with any neighbour within
// the page's dirty gap. The list is fixed size, so a live frame allocates
// nothing.
func (d *Damage) Add(r image.Rectangle) {
	if r.Empty() {
		return
	}

	if d.union.Empty() {
		d.union = r
	} else {
		d.union = d.union.Union(r)
	}

	if d.Overflow {
		d.Rects[0] = d.union

		return
	}

	d.Rects[d.N] = r
	d.N++
	d.N = len(dirty.CoalesceRects(d.Rects[:d.N], dirty.Gap))
	if d.N == len(d.Rects) {
		d.Rects[0] = d.union
		d.N = 1
		d.Overflow = true
	}
}

// Bounds is the one box covering every rectangle added, including ones that
// coalescing merged away.
func (d *Damage) Bounds() image.Rectangle {
	return d.union
}

// Empty reports that the step changed nothing.
func (d *Damage) Empty() bool {
	return d.union.Empty()
}
