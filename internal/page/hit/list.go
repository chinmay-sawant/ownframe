package hit

import "github.com/chinmay-sawant/blinkless/layout"

// Item is one box in the hit list, with the space its rectangle lives in.
// A box the window pins to the viewport is tested in viewport coordinates,
// because the window paints it there while the content below scrolls under it.
type Item struct {
	Box   layout.Box
	Fixed bool
}

// List builds the hit list in document order, so the innermost box under
// a point still wins inside one layer. A box the window pins to the viewport
// is marked, because the window paints that layer last and a click has to
// land on the box the user sees on top.
func List(boxes, fixed []layout.Box) []Item {
	out := make([]Item, 0, len(boxes))
	if len(fixed) == 0 {
		for _, b := range boxes {
			out = append(out, Item{Box: b})
		}

		return out
	}

	roots := make(map[key]bool, len(fixed))
	for _, b := range fixed {
		roots[keyOf(b)] = true
	}
	for i, b := range boxes {
		out = append(out, Item{Box: b, Fixed: inFixed(boxes, i, roots)})
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
