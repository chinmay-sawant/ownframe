package page

import "github.com/chinmay-sawant/blinkless/layout"

// diffDisplay returns the regions covering the operations that differ
// between prev and next. full reports a difference too scattered or too
// large to merge, and changed counts the operations that differ. A change
// stays partial while its bounding box is under a third of the frame and
// its merged rect list fits maxDirtyRects.
func diffDisplay(prev, next *Display) (dirtySet, bool, int) {
	if prev.Width != next.Width || prev.Height != next.Height {
		return dirtySet{}, true, len(next.Ops)
	}

	if len(prev.Ops) != len(next.Ops) {
		return diffByPosition(prev, next)
	}

	d := newDirtySet(next.Width, next.Height)
	changed := 0

	for i := range prev.Ops {
		if sameOp(&prev.Ops[i], &next.Ops[i]) {
			continue
		}

		changed++
		d.add(opBounds(&prev.Ops[i], prev).Union(opBounds(&next.Ops[i], next)))
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

func keyOf(op *DisplayOp) opKey {
	return opKey{op.Kind, op.X, op.Y, op.W, op.H}
}
