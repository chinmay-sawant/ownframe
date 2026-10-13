package window

// zoom returns the page zoom: the pinch scale times the keyboard or wheel
// zoom, clamped to a usable range. A shell that never zoomed reads 1.
func (s *shell) zoom() float64 {
	if s.viewLocked() {
		return 1
	}

	return clampZoom(s.fingers.zoomOr1() * s.zoomLevel())
}

// zoomLevel returns the keyboard or wheel zoom.
func (s *shell) zoomLevel() float64 {
	if s.pageZoom <= 0 {
		return 1
	}

	return s.pageZoom
}

// zoomBy scales the page zoom by factor.
func (s *shell) zoomBy(factor float64) {
	if s.viewLocked() || factor <= 0 {
		return
	}

	s.pageZoom = clampZoom(s.zoomLevel() * factor)
}

// zoomReset returns the page zoom to 1.
func (s *shell) zoomReset() {
	s.pageZoom = 1
}

// clampZoom keeps a zoom inside the pinch range.
func clampZoom(z float64) float64 {
	if z < minZoom {
		return minZoom
	}

	if z > maxZoom {
		return maxZoom
	}

	return z
}

// contentAt maps a window point into the page for the current scroll and
// pinch zoom.
func (s *shell) contentAt(x, y, frameW, frameH int) (float64, float64) {
	if s.viewLocked() {
		scale, ox, oy := s.fit()
		return (float64(x) - ox) / scale, (float64(y) - oy) / scale
	}

	viewW, viewH := s.viewSize()

	return contentPointZoom(
		x, y, s.scrollX, s.scrollY, s.zoom(), s.stretched(),
		frameW, frameH, viewW, viewH,
	)
}

// contentPointZoom maps a cursor into the painted page and divides out the
// pinch zoom, so a click stays on the glyph under the finger.
func contentPointZoom(x, y, scrollX, scrollY int, zoom float64, stretched bool, frameW, frameH, screenW, screenH int) (float64, float64) {
	px, py := contentPoint(x, y, scrollX, scrollY, stretched, frameW, frameH, screenW, screenH)
	if zoom <= 0 {
		zoom = 1
	}

	return px / zoom, py / zoom
}
