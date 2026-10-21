package page

// boxKey identifies a box inside the page's box list. An anonymous box has
// no id to match on, so its tag, its parent, and its rectangle name it.
type boxKey struct {
	id     string
	tag    string
	parent int
	x, y   float64
	w, h   float64
}

// boxKeyOf returns the key of a box.
func boxKeyOf(b Box) boxKey {
	return boxKey{
		id:     b.ID,
		tag:    b.Tag,
		parent: b.ParentIndex,
		x:      b.X,
		y:      b.Y,
		w:      b.W,
		h:      b.H,
	}
}

// pinnedBox reports whether the box is painted with the fixed layer: it is
// the fixed box itself, or it sits inside one, because the engine paints
// the whole subtree of a fixed element with it. The window paints that
// layer last, so such a box covers the content below it.
func pinnedBox(boxes []Box, index int, pinned map[boxKey]bool) bool {
	for i := index; i >= 0 && i < len(boxes); {
		if pinned[boxKeyOf(boxes[i])] {
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

// insideBox reports whether the box is the other box or sits inside it.
func insideBox(boxes []Box, box, other int) bool {
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

// aboveBox reports whether the box is the other box or an ancestor of it.
func aboveBox(boxes []Box, box, other int) bool {
	return insideBox(boxes, other, box)
}
