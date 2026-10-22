package window

// frameMarker is a ticking page that can report dirty rects.
type frameMarker interface {
	FrameDirty() bool
}

// captureFrameOptIn records the page's frame-dirty flag for this Update.
// Ebiten can call Draw without an Update in between, and Page.FrameDirty
// clears the flag when it is read, so the window must read it once per tick.
func (s *shell) captureFrameOptIn() {
	if marker, ok := s.app.(frameMarker); ok {
		s.frameOptIn = marker.FrameDirty()
	}
}

// fullReplay reports a ticking page that cannot report dirty rects.
// A Page (frameMarker) stays on the buffer path: Tick marks mutated ops
// itself, and UseFrameDirty is an alias. A no-op tick leaves TakeDirty
// empty so idle park can sleep. A ticker without FrameDirty still replays
// in full.
func (s *shell) fullReplay() bool {
	if !s.isTicking() {
		return false
	}

	_, ok := s.app.(frameMarker)

	return !ok
}
