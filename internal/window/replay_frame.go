package window

// frameMarker is a ticking page that can opt into the partial buffer.
type frameMarker interface {
	FrameDirty() bool
}

// captureFrameOptIn records the page's frame-dirty opt-in for this Update.
// Ebiten can call Draw without an Update in between, and Page.FrameDirty
// clears the flag when it is read, so the window must read it once per tick.
// Reading it in the draw instead turns every update-less draw into a full
// replay of the whole display list, which on a page with many rounded
// operations is tens of megabytes of vertices per frame.
func (s *shell) captureFrameOptIn() {
	if marker, ok := s.app.(frameMarker); ok {
		s.frameOptIn = marker.FrameDirty()
	}
}

// fullReplay reports a ticking page that did not opt into a dirty rect.
// UseFrameDirty keeps that page on the buffer path.
func (s *shell) fullReplay() bool {
	if !s.isTicking() {
		return false
	}

	if _, ok := s.app.(frameMarker); !ok {
		return true
	}

	return !s.frameOptIn
}
