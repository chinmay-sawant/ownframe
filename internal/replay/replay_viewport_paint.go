package replay

import (
	"image"

	"github.com/chinmay-sawant/blinkless/layout"
	"github.com/hajimehoshi/ebiten/v2"
)

// paintViewportPlan replays the plan for several page-space rects in one pass.
// rects share the picker's translation decision; the picker still decides per
// op. Candidates come from the tile buckets once, sorted into paint order, so
// a rect that spans tiles paints once and several rects share one plan walk.
// The precise ink check uses the viewport union when its area is below the
// threshold and per-rect checks otherwise.
func (renderer *Renderer) paintViewportPlan(dst *ebiten.Image, d *layout.Display, p *chunkPlan, rects []image.Rectangle, pick opPicker) {
	if dst == nil || d == nil || p == nil || len(rects) == 0 {
		return
	}

	kept := rects[:0]

	for _, r := range rects {
		if clipped := r.Intersect(p.content); !clipped.Empty() {
			kept = append(kept, clipped)
		}
	}

	rects = kept

	if len(rects) == 0 {
		return
	}

	union, useUnion := viewportUnion(rects)
	ppt := d.PointsPerPixel

	seen := make([]int, len(p.items))
	var cands []int

	const viewportSeenGen = 1

	if p.tiles != nil {
		for _, r := range rects {
			cands = p.tiles.planTilesForRect(r, seen, viewportSeenGen, cands)
		}

		sortPlanCandidates(cands)

		// Deduplicate after gathering across rects: buckets append per rect,
		// so the same item may appear twice before the sort.
		dedup := cands[:0]

		for i, at := range cands {
			if i == 0 || at != cands[i-1] {
				dedup = append(dedup, at)
			}
		}

		cands = dedup
	} else {
		cands = cands[:0]
		for i := range p.items {
			cands = append(cands, i)
		}
	}

	touches := func(op *layout.DisplayOp) bool {
		return opTouchesAny(op, ppt, rects, union, useUnion)
	}

	for _, at := range cands {
		it := &p.items[at]

		switch {
		case it.group != nil:
			if !groupTouchesAny(it.group, ppt, rects, union, useUnion) {
				continue
			}

			renderer.paintGroup(dst, d, it.group, pickWithTouches(pick, touches))
		case it.run != nil:
			renderer.paintTileRun(dst, d, p, it.run, pickWithTouches(pick, touches))
		case it.op != nil:
			if !touches(it.op) {
				continue
			}

			if dx, dy, ok := pick(it.op); ok {
				countOp()
				renderer.drawOp(dst, it.op, dx, dy)
			}
		}
	}
}

// pickWithTouches wraps a picker with a precise ink precheck, so groups and
// runs skip ops outside the viewport rects in the same single pass.
func pickWithTouches(pick opPicker, touches func(*layout.DisplayOp) bool) opPicker {
	return func(op *layout.DisplayOp) (float64, float64, bool) {
		if !touches(op) {
			return 0, 0, false
		}

		return pick(op)
	}
}

// groupTouchesAny reports whether a group's page box meets the viewport set.
// Unbounded groups always touch.
func groupTouchesAny(n *groupNode, ppt float64, rects []image.Rectangle, union image.Rectangle, useUnion bool) bool {
	box, ok := nodeBox(n, ppt)
	if !ok {
		return true
	}

	if box.Empty() {
		return false
	}

	if useUnion {
		return !box.Intersect(union).Empty()
	}

	for _, r := range rects {
		if !box.Intersect(r).Empty() {
			return true
		}
	}

	return false
}
