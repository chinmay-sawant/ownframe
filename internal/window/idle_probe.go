package window

// idleCleanFrame reports that the frame just drawn needed no raster work:
// the page replayed into the persistent buffer at one to one with no
// repaint and no overlay. The pacing wrapper calls it after Draw, and it
// only reads state. Every condition that may have taken a full-replay or
// scaled branch reads as dirty, so a missed idle frame is cheaper than a
// wrongly throttled one. Pinned chrome stays on the tile path; a later
// WantFrameAt still wakes an animating bar. Overlay chrome is the
// wrapper's call: an unchanged set of chrome rects paints the same
// pixels and does not need the vsync loop.
func (s *shell) idleCleanFrame() bool {
	if s.fallback || s.dev.on || s.display == nil {
		return false
	}

	if s.partial.mode != repaintBlit || !s.partial.live || s.partial.whole {
		return false
	}

	if s.partial.buf == nil {
		return false
	}

	if s.viewLocked() || s.bitmapView.img != nil || s.stretched() || s.zoom() != 1 {
		return false
	}

	if s.fullReplay() || !s.bufferOrderSafe() {
		return false
	}

	_, canTake := s.damageTaker()

	return canTake
}
