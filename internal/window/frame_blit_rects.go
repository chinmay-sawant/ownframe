package window

import (
	"image"
)

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
	if len(screen) == 0 {
		return nil
	}

	out := make([]image.Rectangle, 0, len(screen))
	for _, rect := range screen {
		out = append(out, rect.Add(scroll))
	}

	return out
}
