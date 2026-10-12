package window

import (
	"image"

	"github.com/hajimehoshi/ebiten/v2"
)

// maxBlitRects caps one frame's screen blits; past it blitPlan collapses to
// one whole-buffer copy.
const maxBlitRects = 8

// blitPlan decides one frame's screen blits. full means the caller painted the
// whole screen, so no blits are needed. Rects are in buffer coordinates, which
// blitBuffer offsets by the scroll. A repaintBuffer plan means the screen
// repaints in full; repaintBlit contributes nothing. Overlapping rects merge;
// past the cap one union covers them all.
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

// blitBuffer copies the planned rects from the partial buffer to dst, offset
// by the scroll, skipping empties. whole copies the whole buffer instead: a
// scroll change moved the picture, and one copy covers every plan rect and
// the chrome restore. The copy replaces pixels, so a transparent page erases.
func (s *shell) blitBuffer(dst *ebiten.Image, plans []repaintPlan, chrome []image.Rectangle, whole bool) {
	buf := s.partial.buf
	if buf == nil || dst == nil {
		return
	}

	var rects []image.Rectangle

	if whole {
		rects = []image.Rectangle{buf.Bounds()}
	} else if full, planned := blitPlan(plans, chrome); full {
		return
	} else {
		rects = planned
	}

	for _, rect := range rects {
		s.blitRect(dst, buf, rect)
	}
}
