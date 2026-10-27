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
	return blitPlanBuf(nil, plans, chrome)
}

func blitPlanBuf(buf []image.Rectangle, plans []repaintPlan, chrome []image.Rectangle) (full bool, rects []image.Rectangle) {
	rects = buf[:0]
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
		if cap(rects) > 0 {
			u := unionRects(rects)
			rects = rects[:1]
			rects[0] = u
			return false, rects
		}

		return false, []image.Rectangle{unionRects(rects)}
	}

	return false, rects
}

// mergeBlitRects unions overlapping rects until no pair overlaps. A frame
// brings at most the plan cap plus the chrome rects, so the restarting scan
// stays cheap.
func mergeBlitRects(rects []image.Rectangle) []image.Rectangle {
	for changed := true; changed; {
		changed = false

		for i := 0; i < len(rects) && !changed; i++ {
			for j := i + 1; j < len(rects); j++ {
				if rects[i].Overlaps(rects[j]) {
					rects[i] = rects[i].Union(rects[j])
					rects = append(rects[:j], rects[j+1:]...)
					changed = true

					break
				}
			}
		}
	}

	return rects
}

// unionRects is the smallest rect covering every rect. blitPlan uses it past
// the cap: one copy of the whole damaged area.
func unionRects(rects []image.Rectangle) image.Rectangle {
	if len(rects) == 0 {
		return image.Rectangle{}
	}

	box := rects[0]
	for _, rect := range rects[1:] {
		box = box.Union(rect)
	}

	return box
}

// chromeRects moves the previous frame's chrome rects from screen space into
// buffer space. The current scroll says which buffer pixels belong under the
// screen rect now, so a restore repairs stale chrome even after a scroll.
func chromeRects(screen []image.Rectangle, scroll image.Point) []image.Rectangle {
	return chromeRectsBuf(nil, screen, scroll)
}

func chromeRectsBuf(buf []image.Rectangle, screen []image.Rectangle, scroll image.Point) []image.Rectangle {
	if len(screen) == 0 {
		return nil
	}

	out := buf[:0]
	for _, rect := range screen {
		out = append(out, rect.Add(scroll))
	}

	return out
}
