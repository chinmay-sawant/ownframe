package hit

import "github.com/chinmay-sawant/blinkless/layout"

// Item is one box in the hit list, with the space its rectangle lives in.
// A box the window pins to the viewport is tested in viewport coordinates,
// because the window paints it there while the content below scrolls under it.
type Item struct {
	Box   layout.Box
	Fixed bool
}

// maxLinearFixed caps the indexed fixed-root path below. At most that many
// pinned boxes resolve by linear scan; more fall back to the keyed map, so a
// page that pins everything never pays a quadratic scan.
const maxLinearFixed = 32

// List rebuilds the hit list in out's storage in document order, so the
// innermost box under a point still wins inside one layer. A viewport-pinned
// box is marked because the window paints that layer last and a click has
// to land on the box the user sees on top.
func List(out []Item, boxes, fixed []layout.Box) []Item {
	if cap(out) < len(boxes) {
		out = make([]Item, len(boxes))
	} else {
		if len(out) > len(boxes) {
			clear(out[len(boxes):])
		}
		out = out[:len(boxes)]
	}
	if len(fixed) == 0 {
		for i := range boxes {
			out[i] = Item{Box: boxes[i]}
		}

		return out
	}
	if len(fixed) > maxLinearFixed {
		return listKeyed(out, boxes, fixed)
	}

	roots := fixedRootSet(boxes, fixed)
	for i := range boxes {
		out[i] = Item{Box: boxes[i], Fixed: inFixedSet(boxes, i, roots)}
	}

	return out
}

// fixedRootSet resolves the pinned boxes to their indices in boxes. A fixed
// box is always an element of the box list it was derived from, so a miss
// only means a stale caller and is skipped the way a dead map key is.
func fixedRootSet(boxes, fixed []layout.Box) []int {
	roots := make([]int, 0, len(fixed))
	for _, fb := range fixed {
		if idx := indexOfBox(boxes, fb); idx >= 0 {
			roots = append(roots, idx)
		}
	}
	return roots
}

// indexOfBox returns the first index whose box names the same element with
// the same rectangle as want: the id, the tag, the parent link, and the
// rect. Those are exactly the fields the keyed path hashes on, so both
// paths resolve the same roots.
func indexOfBox(boxes []layout.Box, want layout.Box) int {
	for i := range boxes {
		if sameHitBox(boxes[i], want) {
			return i
		}
	}
	return -1
}

// sameHitBox compares the fields a hit-test box keeps across the hit list
// and the display: the id, the tag, the parent link, and the rectangle.
func sameHitBox(a, b layout.Box) bool {
	return a.ID == b.ID && a.Tag == b.Tag && a.ParentIndex == b.ParentIndex &&
		a.X == b.X && a.Y == b.Y && a.W == b.W && a.H == b.H
}

// inFixedSet reports whether the box is pinned with the fixed layer, the
// way inFixed does: it is a fixed root itself, or it sits inside one. The
// walk and its guards match inFixed, with index membership instead of a
// keyed lookup per ancestor.
func inFixedSet(boxes []layout.Box, index int, roots []int) bool {
	for i := index; i >= 0 && i < len(boxes); {
		for _, r := range roots {
			if i == r {
				return true
			}
		}
		next := boxes[i].ParentIndex
		if next >= i {
			return false
		}
		i = next
	}

	return false
}

// listKeyed is the keyed fixed-marking path for many pinned boxes: one map
// of fixed keys, then one ancestor climb per box. The indexed path above
// wins while the pinned set is small because it hashes nothing.
func listKeyed(out []Item, boxes, fixed []layout.Box) []Item {
	roots := make(map[key]bool, len(fixed))
	for _, b := range fixed {
		roots[keyOf(b)] = true
	}
	for i := range boxes {
		out[i] = Item{Box: boxes[i], Fixed: inFixed(boxes, i, roots)}
	}

	return out
}

// key identifies a box inside the page's box list. An anonymous box has
// no id to match on, so its tag, its parent, and its rectangle name it.
type key struct {
	id     string
	tag    string
	parent int
	x, y   float64
	w, h   float64
}

func keyOf(b layout.Box) key {
	return key{
		id:     b.ID,
		tag:    b.Tag,
		parent: b.ParentIndex,
		x:      b.X,
		y:      b.Y,
		w:      b.W,
		h:      b.H,
	}
}

// inFixed reports whether the box is painted with the fixed layer: it is
// the fixed box itself, or it sits inside one, because the engine paints
// the whole subtree of a fixed element with it. The window paints that
// layer last, so such a box covers the content below it.
func inFixed(boxes []layout.Box, index int, roots map[key]bool) bool {
	for i := index; i >= 0 && i < len(boxes); {
		if roots[keyOf(boxes[i])] {
			return true
		}

		next := boxes[i].ParentIndex
		if next >= i {
			return false
		}
		i = next
	}

	return false
}

// inside reports whether box is other or sits inside it.
func inside(boxes []layout.Box, box, other int) bool {
	for i := box; i >= 0 && i < len(boxes); {
		if i == other {
			return true
		}

		next := boxes[i].ParentIndex
		if next >= i {
			return false
		}
		i = next
	}

	return false
}

// above reports whether box is other or an ancestor of it.
func above(boxes []layout.Box, box, other int) bool {
	return inside(boxes, other, box)
}
