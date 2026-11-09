package dirty

import (
	"image"

	"github.com/chinmay-sawant/blinkless/layout"
)

// labelGap is how near a label must sit to a toggle to repaint with it.
// It is wider than Gap: a label is not a neighbouring damage rectangle.
const labelGap = 12

// boxRound nudges a box's far edges to the nearest whole CSS pixel.
const boxRound = 0.5

// BoxRect is one box in whole CSS pixels. The origin is truncated; the
// far edges gain boxRound before the truncation.
func BoxRect(b layout.Box) image.Rectangle {
	return image.Rect(int(b.X), int(b.Y), int(b.X+b.W+boxRound), int(b.Y+b.H+boxRound))
}

// GrowBox covers the boxes inside r, so a nested span repaints with its
// parent. An empty r is returned unchanged.
//
// GrowBox can only return r unchanged for a non-empty r: it unions only
// boxes already inside r, and a box inside r adds nothing to the union.
// The subtree walk below supersedes it: an overflowing descendant sits
// outside r, so geometry alone never covers it, while the index walk
// does. GrowBox stays as the legacy reference the tests compare against.
func GrowBox(r image.Rectangle, boxes []layout.Box) image.Rectangle {
	if r.Empty() {
		return r
	}

	for _, b := range boxes {
		inner := BoxRect(b)
		if !inner.Empty() && inner.In(r) {
			r = r.Union(inner)
		}
	}

	return r
}

// KidIndex maps each element box to its element children, so a mark
// walks only the changed element's subtree instead of scanning every
// box. The rects snapshot is taken at build time, so the index never
// reads a box list that a later commit replaced.
type KidIndex struct {
	byID  map[string]int
	head  []int
	link  []int
	rects []image.Rectangle
}

// BuildKidIndex builds the parent-to-children index for boxes. A box
// whose parent link is missing or out of range is treated as a root.
// An empty box list builds an empty index that reports every id
// unknown.
func BuildKidIndex(boxes []layout.Box) KidIndex {
	k := KidIndex{
		byID:  make(map[string]int, len(boxes)),
		head:  make([]int, len(boxes)),
		link:  make([]int, len(boxes)),
		rects: make([]image.Rectangle, len(boxes)),
	}
	for i := range boxes {
		k.byID[boxes[i].ID] = i
		k.head[i] = -1
		k.link[i] = -1
		k.rects[i] = BoxRect(boxes[i])
	}
	for i := range boxes {
		pi := boxes[i].ParentIndex
		if pi < 0 || pi >= len(boxes) || pi == i {
			// A missing, out-of-range, or self parent is a root. The self
			// case would otherwise link the box to itself and make the
			// subtree walk cycle forever.
			continue
		}
		k.link[i] = k.head[pi]
		k.head[pi] = i
	}
	return k
}

// SameParents reports whether a and b carry the same parent links in the same
// order. Paired with an unchanged id sequence it proves the child lists a
// KidIndex built from a still describe b, so a commit only needs to rewrite
// the rects snapshot.
func SameParents(a, b []layout.Box) bool {
	if len(a) != len(b) {
		return false
	}
	for i := range a {
		if a[i].ParentIndex != b[i].ParentIndex {
			return false
		}
	}
	return true
}

// UpdateRects rewrites the rects snapshot from boxes for an index whose id
// sequence and parent links still describe boxes. The map and the child links
// stay as they are. A length mismatch leaves the index untouched; the caller
// proves the match with SameIDs and SameParents.
func (k KidIndex) UpdateRects(boxes []layout.Box) {
	if len(k.rects) != len(boxes) {
		return
	}
	for i := range boxes {
		k.rects[i] = BoxRect(boxes[i])
	}
}

// Rect returns the union of the id's box and every descendant's box, so
// an overflowing child repaints with its parent. ok is false when the
// index has no box for id, and the caller falls back to the id's own
// box. A contained subtree unions exactly to the id's own box, matching
// GrowBox.
func (k KidIndex) Rect(id string) (r image.Rectangle, ok bool) {
	if k.byID == nil {
		return image.Rectangle{}, false
	}
	root, found := k.byID[id]
	if !found {
		return image.Rectangle{}, false
	}
	r = k.rects[root]
	stack := []int{root}
	visited := 0
	for len(stack) > 0 {
		n := stack[len(stack)-1]
		stack = stack[:len(stack)-1]
		visited++
		if visited > len(k.rects) {
			// A malformed parent cycle would otherwise walk forever.
			return r, true
		}
		for c := k.head[n]; c != -1; c = k.link[c] {
			r = r.Union(k.rects[c])
			stack = append(stack, c)
		}
	}
	return r, true
}

// ToggleHalo is the margin around a checkbox or radio that covers an
// inline label the layout did not box separately.
func ToggleHalo(ctl image.Rectangle) image.Rectangle {
	return ctl.Inset(-labelGap)
}

// LabelNear reports b's box when b is a label within labelGap of the
// control. Otherwise it reports false.
func LabelNear(ctl image.Rectangle, b layout.Box) (image.Rectangle, bool) {
	if b.Tag != "label" {
		return image.Rectangle{}, false
	}

	r := BoxRect(b)
	if !touches(ctl, r, labelGap) {
		return image.Rectangle{}, false
	}

	return r, true
}
