package dirty

import "image"

// CoalesceRects merges rects that meet within gap and returns the shorter
// list. It compacts in place on the given backing array.
func CoalesceRects(rects []image.Rectangle, gap int) []image.Rectangle {
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

// clampRects pads every rect by Pad, confines it to content, and drops
// the empties. It compacts in place on the given backing array.
func clampRects(rects []image.Rectangle, content image.Rectangle) []image.Rectangle {
	out := rects[:0]
	for _, r := range rects {
		r = r.Inset(-Pad).Intersect(content)
		if r.Empty() {
			continue
		}

		out = append(out, r)
	}

	return out
}

// BoundingBox is the one rect covering every rect. The caller passes a
// non-empty list.
func BoundingBox(rects []image.Rectangle) image.Rectangle {
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

func touches(a, b image.Rectangle, gap int) bool {
	return !a.Inset(-gap).Intersect(b).Empty()
}
