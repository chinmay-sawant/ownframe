package page

import "image"

// dirtySet collects the rects one display diff found and decides whether
// they stay a partial repaint. It keeps at most maxDirtyRects rects: past
// the cap it unions the closest pair. The bounding box must stay under a
// third of the frame, so a distant outlier still forces a full frame.
type dirtySet struct {
	rects [maxDirtyRects]image.Rectangle
	n     int
	union image.Rectangle
	full  bool
	frame image.Rectangle
}

// newDirtySet prepares an empty set for a w x h frame.
func newDirtySet(w, h int) dirtySet {
	return dirtySet{frame: image.Rect(0, 0, w, h)}
}

// add inserts one changed op's bounds. Once the bounding box crosses a
// third of the frame the set turns full and later adds are ignored.
func (d *dirtySet) add(r image.Rectangle) {
	if d.full || r.Empty() {
		return
	}

	u := d.union.Union(r)
	if u.Dx()*u.Dy() > d.frame.Dx()*d.frame.Dy()/3 {
		d.full = true

		return
	}

	d.union = u
	if d.n == len(d.rects) {
		d.mergeClosest()
	}

	d.rects[d.n] = r
	d.n++
	d.coalesce()
}

// coalesce merges every pair that meets within dirtyGap.
func (d *dirtySet) coalesce() {
	d.n = len(coalesceRects(d.rects[:d.n], dirtyGap))
}

// mergeClosest unions the pair whose union is smallest, so the cap costs
// the least extra area.
func (d *dirtySet) mergeClosest() {
	rs := d.rects[:d.n]
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
	d.n--
}

// bounds is the bounding box of every rect added.
func (d *dirtySet) bounds() image.Rectangle {
	return d.union
}

// unionArea is the area of the box covering a and b.
func unionArea(a, b image.Rectangle) int {
	u := a.Union(b)

	return u.Dx() * u.Dy()
}
