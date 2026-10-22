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
		if right := int(math.Ceil(b.X + b.W)); right > w {
			w = right
		}

		if bottom := int(math.Ceil(b.Y + b.H)); bottom > h {
			h = bottom
		}
	}

	if gen != 0 {
		s.contentW, s.contentH, s.contentGen = w, h, gen
	}

	return w, h
}
