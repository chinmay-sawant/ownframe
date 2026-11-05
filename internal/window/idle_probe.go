package window

// idleCleanFrame reports that the frame just drawn needed no raster work:
// the page replayed into the persistent buffer at one to one with no
// repaint and no overlay, or a stable viewport blit reused its image.
// The pacing wrapper calls it after Draw, and it only reads state. Every
// condition that may have taken a full-replay or scaled branch reads as
// dirty, so a missed idle frame is cheaper than a wrongly throttled one.
// Pinned chrome stays on the tile path; a later WantFrameAt still wakes an
// animating bar. Overlay chrome is the wrapper's call: an unchanged set of
// chrome rects paints the same pixels and does not need the active drawing loop.
// A quiet ticker idles on either path: the tile path already does, and the
// viewport path does when the damage taker stayed clean and the tick set no
// FrameDirty. A ticker that cannot report damage stays on the frame clock.
func (s *shell) idleCleanFrame() bool {
	if s.fallback || s.display == nil {
		return false
	}

	if s.viewLocked() || s.bitmapView.img != nil || s.stretched() || s.zoom() != 1 {
		return false
	}

	if s.fullReplay() || !s.bufferOrderSafe() {
		return false
	}

	if _, canTake := s.damageTaker(); !canTake {
		return false
	}
	if s.renderer.HasBackdropFilter(s.display) {
		return s.partial.mode == repaintBlit && s.partial.live && !s.partial.whole && s.partial.backdropClean
	}

	// A retained layer is composited after the background blit, so the frame
	// is not clean even when the page reports no damage: the layer may move.
	if s.display != nil && s.renderer.LayersActive(s.display) {
		return false
	}

	if s.viewport.valid && s.viewport.buf != nil {
		// blitOnly means the damage taker reported nothing this frame;
		// frameOptIn means the tick set FrameDirty (a mutating tick or
		// UseFrameDirty). Both clean idles. A ticker without FrameDirty
		// cannot report quiet, so it stays dirty.
		if s.isTicking() {
			if _, ok := s.app.(frameMarker); !ok {
				return false
			}

			return s.viewport.blitOnly && !s.frameOptIn
		}

		return s.viewport.blitOnly
	}

	if s.partial.mode != repaintBlit || !s.partial.live || s.partial.whole {
		return false
	}

	return s.partial.buf != nil
}
