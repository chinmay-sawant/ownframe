package dirty

import "github.com/chinmay-sawant/blinkless/layout"

// Diff returns the regions covering the operations that differ between
// prev and next. full reports a difference too scattered or too large to
// merge, and changed counts the operations that differ. A change stays
// partial while its bounding box is under a third of the frame and its
// merged rectangle list fits MaxRects.
func Diff(prev, next *layout.Display) (Set, bool, int) {
	if prev.Width != next.Width || prev.Height != next.Height {
		return Set{}, true, len(next.Ops)
	}

	if len(prev.Ops) != len(next.Ops) {
		return diffByPosition(prev, next)
	}

	d := newSet(next.Width, next.Height)
	changed := 0

	for i := range prev.Ops {
		if SameOp(&prev.Ops[i], &next.Ops[i]) {
			continue
		}

		changed++
		d.add(OpBounds(&prev.Ops[i], prev).Union(OpBounds(&next.Ops[i], next)))
		if d.full {
			return d, true, changed
		}
	}

	return d, false, changed
}

// diffByPosition pairs operations by kind and geometry when the two lists
// differ in length. An operation with no partner dirties its bounds, and
// changed counts every unpaired or repainted operation.
func diffByPosition(prev, next *layout.Display) (Set, bool, int) {
	queue := make(map[opKey][]int, len(next.Ops))
	for i := range next.Ops {
		k := keyOf(&next.Ops[i])
		queue[k] = append(queue[k], i)
	}

	seen := make([]bool, len(next.Ops))
	d := newSet(next.Width, next.Height)
	changed := 0

	for i := range prev.Ops {
		k := keyOf(&prev.Ops[i])
		list := queue[k]

		if len(list) == 0 {
			changed++
			d.Removed++
			d.add(OpBounds(&prev.Ops[i], prev))
			if d.full {
				return d, true, changed
			}

			continue
		}

		j := list[0]
		queue[k] = list[1:]
		seen[j] = true

		if SameOp(&prev.Ops[i], &next.Ops[j]) {
			continue
		}

		changed++
		d.add(OpBounds(&prev.Ops[i], prev).Union(OpBounds(&next.Ops[j], next)))
		if d.full {
			return d, true, changed
		}
	}

	for j := range next.Ops {
		if seen[j] {
			continue
		}

		changed++
		d.Added++
		d.add(OpBounds(&next.Ops[j], next))
		if d.full {
			return d, true, changed
		}
	}

	return d, false, changed
}

// opKey identifies an operation by its paint kind and geometry.
type opKey struct {
	kind       layout.DisplayKind
	x, y, w, h float64
}

func keyOf(op *layout.DisplayOp) opKey {
	return opKey{kind: op.Kind, x: op.X, y: op.Y, w: op.W, h: op.H}
}
