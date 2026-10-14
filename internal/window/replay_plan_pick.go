package window

import "image"

// planRepaints decides the frame's tile work for content contentW x contentH.
// The content can outgrow the canvas when a fixed root's children overflow, so
// the grid and the full rect use the content. A frame with no damage blits
// when the grid holds this generation and rebuilds it otherwise. A
// full-content rect, a list longer than maxRepaintRects, or a rect that clamps
// away rebuilds the grid. Otherwise every rect gets its own repaint; all
// plans share one mode. clipX is the first page x the devtools dock hides:
// damage right of it needs no pixels, and a clip that grew repaints the strip
// the dock hid.
func planRepaints(contentW, contentH int, st partialState, gen uint64, rects []image.Rectangle, ok bool, clipX int) []repaintPlan {
	full := image.Rect(0, 0, contentW, contentH)
	rebuild := []repaintPlan{{mode: repaintBuffer, rect: full}}
	if full.Empty() || !st.gridCurrent(contentW, contentH) {
		return rebuild
	}

	strip := st.dockStrip(clipX, contentH)
	if !ok || len(rects) == 0 {
		if st.gen != gen {
			return rebuild
		}

		if strip.Empty() {
			return []repaintPlan{{mode: repaintBlit, rect: full}}
		}

		return []repaintPlan{{mode: repaintRect, rect: strip}}
	}

	plans := make([]repaintPlan, 0, len(rects)+1)
	for _, rect := range rects {
		inside := rect.Intersect(full)
		if inside.Empty() {
			return rebuild
		}

		visible := clipDockX(inside, clipX)
		if visible.Empty() {
			continue
		}

		if visible == full {
			return rebuild
		}

		plans = append(plans, repaintPlan{mode: repaintRect, rect: visible})
	}

	if !strip.Empty() {
		plans = append(plans, repaintPlan{mode: repaintRect, rect: strip})
	}

	if len(plans) == 0 {
		return []repaintPlan{{mode: repaintBlit, rect: full}}
	}

	if len(plans) > maxRepaintRects {
		return rebuild
	}

	return plans
}
