package window

import (
	"image"
)

// maxBlitRects caps one frame's screen blits; past it blitPlan collapses to
// one whole-buffer copy.
const maxBlitRects = 8

// blitPlan decides one frame's screen blits. full means the caller painted the
// whole screen, so no blits are needed. Rects are in buffer coordinates, which
// blitTiles offsets by the scroll and trims to the dock's visible page. A
// repaintBuffer plan means the screen repaints in full; repaintBlit
// contributes nothing. Overlapping rects merge; past the cap one union covers
// them all.
func blitPlan(plans []repaintPlan, chrome []image.Rectangle) (full bool, rects []image.Rectangle) {
	for _, plan := range plans {
		if plan.mode == repaintBuffer {
			return true, nil
		}

		if plan.mode == repaintRect && !plan.rect.Empty() {
			rects = append(rects, plan.rect)
		}
	}

	for _, rect := range chrome {
		if !rect.Empty() {
			rects = append(rects, rect)
		}
	}

	rects = mergeBlitRects(rects)
	if len(rects) > maxBlitRects {
		return false, []image.Rectangle{unionRects(rects)}
	}

	return false, rects
}
