package window

import "image"

// repaintMode is what one frame does with the persistent buffer.
type repaintMode int

const (
	// repaintFull replays the list straight to the screen.
	repaintFull repaintMode = iota
	// repaintBuffer repaints the whole buffer.
	repaintBuffer
	// repaintRect repaints one dirty rect in the buffer.
	repaintRect
	// repaintBlit leaves the buffer alone. Its pixels are current.
	repaintBlit
)

// repaintPlan is the buffer work for one frame.
type repaintPlan struct {
	mode repaintMode
	rect image.Rectangle
}

// maxRepaintRects caps the dirty rects of one frame. Each rect costs a scan
// of the operation list, so a longer list rebuilds the whole buffer.
const maxRepaintRects = 8

// planRepaints decides the frame's buffer work for content contentW x contentH.
// The content can outgrow the canvas when a fixed root's children overflow, so
// the buffer and the full rect use the content. A frame with no damage blits
// when the buffer holds this generation and rebuilds it otherwise. A
// full-content rect, a list longer than maxRepaintRects, or a rect that clamps
// away rebuilds the buffer. Otherwise every rect gets its own repaint; all
// plans share one mode.
func planRepaints(contentW, contentH int, st partialState, gen uint64, rects []image.Rectangle, ok bool) []repaintPlan {
	full := image.Rect(0, 0, contentW, contentH)
	rebuild := []repaintPlan{{mode: repaintBuffer, rect: full}}
	if full.Empty() || st.buf == nil || st.width != contentW || st.height != contentH {
		return rebuild
	}

	if !ok || len(rects) == 0 {
		if st.gen == gen {
			return []repaintPlan{{mode: repaintBlit, rect: full}}
		}

		return rebuild
	}

	if len(rects) > maxRepaintRects {
		return rebuild
	}

	plans := make([]repaintPlan, 0, len(rects))
	for _, rect := range rects {
		clamped := rect.Intersect(full)
		if clamped.Empty() || clamped == full {
			return rebuild
		}

		plans = append(plans, repaintPlan{mode: repaintRect, rect: clamped})
	}

	return plans
}
