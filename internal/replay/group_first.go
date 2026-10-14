package replay

import "image"

// firstPick returns the translation of the node's first visible op, or ok
// false when nothing in the node paints.
func firstPick(n *groupNode, pick opPicker) ([2]float64, bool) {
	for i := range n.items {
		it := &n.items[i]
		if it.child != nil {
			if base, ok := firstPick(it.child, pick); ok {
				return base, true
			}

			continue
		}

		if dx, dy, ok := pick(it.op); ok {
			return [2]float64{dx, dy}, true
		}
	}

	return [2]float64{}, false
}

// nodeBox is the node's page box in CSS pixels, the union of its ops and
// children. ok is false when any op is unbounded or the node is empty; the
// caller then uses the whole destination.
func nodeBox(n *groupNode, ppt float64) (image.Rectangle, bool) {
	if !n.known {
		n.known = true

		for i := range n.items {
			it := &n.items[i]

			var (
				box image.Rectangle
				ok  bool
			)

			if it.child != nil {
				box, ok = nodeBox(it.child, ppt)
			} else {
				box, ok = opBounds(it.op, ppt)
			}

			if !ok {
				n.unbounded = true
			}

			n.box = n.box.Union(box)
		}
	}

	return n.box, !n.unbounded
}
