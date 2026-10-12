package page

import "image"

// maxDirtyRects caps the damage list. Past the cap the list collapses to
// its bounding box, so one take never pays for an unbounded list.
const maxDirtyRects = 8

// coalesceRects merges rects that meet within gap and returns the shorter
// list. It compacts in place on the given backing array.
func coalesceRects(rects []image.Rectangle, gap int) []image.Rectangle {
	for merged := true; merged; {
		merged = false

		for i := 0; i < len(rects) && !merged; i++ {
			for j := i + 1; j < len(rects); j++ {
				if !touches(rects[i], rects[j], gap) {
					continue
				}

				rects[i] = rects[i].Union(rects[j])
				rects[j] = rects[len(rects)-1]
				rects = rects[:len(rects)-1]
				merged = true

				break
			}
		}
	}

	return rects
}

// clampRects pads every rect by dirtyPad, confines it to content, and drops
// the empties. It compacts in place on the given backing array.
func clampRects(rects []image.Rectangle, content image.Rectangle) []image.Rectangle {
	out := rects[:0]
	for _, r := range rects {
		r = r.Inset(-dirtyPad).Intersect(content)
		if r.Empty() {
			continue
		}

		out = append(out, r)
	}

	return out
}

// boundingBox is the one rect covering every rect. The caller passes a
// non-empty list.
func boundingBox(rects []image.Rectangle) image.Rectangle {
	r := rects[0]
	for _, b := range rects[1:] {
		r = r.Union(b)
	}

	return r
}

// rectsArea sums the areas of the rects the list holds after coalescing,
// which leaves no two of them overlapping.
func rectsArea(rects []image.Rectangle) int {
	area := 0
	for _, r := range rects {
		area += r.Dx() * r.Dy()
	}

	return area
}
