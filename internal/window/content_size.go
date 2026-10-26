package window

import "math"

// contentSize returns the painted page size, including any box that
// overflows the canvas, so a fixed-width child larger than the window is
// scrollable. The size is cached per page generation: boxes change only on
// Redraw, so repeated calls within one frame cost O(1) on large pages.
// Generation zero never caches, so tests driving a fake screen stay fresh.
func (s *shell) contentSize() (int, int) {
	gen := s.app.Generation()
	if gen != 0 && s.contentGen == gen {
		return s.contentW, s.contentH
	}

	w, h := s.canvasSize()

	viewport := s.viewportOperations()
	for _, b := range s.app.Boxes() {
		if viewportBox(b, viewport) {
			continue
		}

		x2, y2 := b.X+b.W, b.Y+b.H

		if b.HasOverflowClip {
			// Paint clipped by an overflow ancestor never extends the
			// scrollable document: intersect the box with the recorded
			// clip instead of the raw border box. A marquee track twice
			// its lane, for example, must not summon a horizontal
			// scrollbar on every page that clips it.
			c := b.OverflowClip
			x1 := math.Max(b.X, c[0])
			y1 := math.Max(b.Y, c[1])
			x2 = math.Min(x2, c[0]+c[2])
			y2 = math.Min(y2, c[1]+c[3])

			if x2 <= x1 || y2 <= y1 {
				continue
			}
		}

		if right := int(math.Ceil(x2)); right > w {
			w = right
		}

		if bottom := int(math.Ceil(y2)); bottom > h {
			h = bottom
		}
	}

	if gen != 0 {
		s.contentW, s.contentH, s.contentGen = w, h, gen
	}

	return w, h
}
