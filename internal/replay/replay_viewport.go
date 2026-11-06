package replay

import (
	"image"

	"github.com/chinmay-sawant/blinkless/layout"
)

// viewportUnionFactor bounds the overdraw a union rect may add before the
// picker tests each rect instead. A union at most half again the summed area
// paints little outside the rects, so one bounds check per op beats one per
// rect; a sprawling union falls back to per-rect checks.
const viewportUnionFactor = 3

// viewportUnion returns the bounding union of rects and whether one bounds
// check against it covers the set. Empty rects are skipped. A single rect
// always uses the union fast path. For several rects, the union wins when its
// area is below the threshold: union*2 <= sum*3, so at most half again the
// precise area paints.
func viewportUnion(rects []image.Rectangle) (image.Rectangle, bool) {
	kept := rects[:0]

	for _, r := range rects {
		if !r.Empty() {
			kept = append(kept, r)
		}
	}

	rects = kept

	if len(rects) == 0 {
		return image.Rectangle{}, false
	}

	if len(rects) == 1 {
		return rects[0], true
	}

	union := rects[0]
	var sum int64

	for _, r := range rects {
		union = union.Union(r)
		sum += int64(r.Dx()) * int64(r.Dy())
	}

	unionArea := int64(union.Dx()) * int64(union.Dy())
	if unionArea*2 <= sum*viewportUnionFactor {
		return union, true
	}

	return union, false
}

// opTouchesAny reports whether an op's painted box meets any rect, with one
// bounds lookup when useUnion holds and one per rect otherwise. An op whose
// ink cannot be bounded always touches, matching opTouches.
func opTouchesAny(op *layout.DisplayOp, ppt float64, rects []image.Rectangle, union image.Rectangle, useUnion bool) bool {
	if useUnion {
		return opTouches(op, ppt, union)
	}

	for _, r := range rects {
		if opTouches(op, ppt, r) {
			return true
		}
	}

	return false
}
