package dirty

import "image"

// Set collects the rectangles one display diff found. It keeps at most
// MaxRects of them: past the cap it unions the closest pair. The bounding
// box must stay under a third of the frame, so a distant outlier still
// forces a full frame. Rects beyond N are unused.
type Set struct {
	Rects   [MaxRects]image.Rectangle
	N       int
	Added   int
	Removed int

	union image.Rectangle
	frame image.Rectangle
	full  bool
}

// Bounds is the box covering every rectangle added, including ones later
// merged away. It is not recomputed from the live list.
func (d *Set) Bounds() image.Rectangle {
	return d.union
}

// newSet prepares an empty set for a w by h frame.
func newSet(w, h int) Set {
	return Set{frame: image.Rect(0, 0, w, h)}
}

// add inserts one changed op's bounds. Once the bounding box crosses a
// third of the frame the set turns full and later adds are ignored.
func (d *Set) add(r image.Rectangle) {
	if d.full || r.Empty() {
		return
	}

	u := d.union.Union(r)
	if u.Dx()*u.Dy() > d.frame.Dx()*d.frame.Dy()/frameThird {
		d.full = true

		return
	}

	d.union = u
	if d.N == len(d.Rects) {
		d.mergeClosest()
	}

	d.Rects[d.N] = r
	d.N++
	d.coalesce()
}

// coalesce merges every pair that meets within Gap.
func (d *Set) coalesce() {
	d.N = len(CoalesceRects(d.Rects[:d.N], Gap))
}

// mergeClosest unions the pair whose union is smallest, so the cap costs
// the least extra area.
func (d *Set) mergeClosest() {
	rs := d.Rects[:d.N]
	bi, bj := 0, 1
	best := unionArea(rs[0], rs[1])

	for i := range rs {
		for j := i + 1; j < len(rs); j++ {
			if a := unionArea(rs[i], rs[j]); a < best {
				best, bi, bj = a, i, j
			}
		}
	}

	rs[bi] = rs[bi].Union(rs[bj])
	rs[bj] = rs[len(rs)-1]
	d.N--
}

// unionArea is the area of the box covering a and b.
func unionArea(a, b image.Rectangle) int {
	u := a.Union(b)

	return u.Dx() * u.Dy()
}
