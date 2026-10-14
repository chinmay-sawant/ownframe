package window

import (
	"image"
	"math"
)

// frameState holds the chrome rects painted on the screen outside the page
// buffer: the devtools overlays, an open menu, each scrollbar thumb, and the
// fallback badge. A partial frame first restores the previous frame's rects
// from the buffer, so closed or moved chrome leaves no stale pixels. Rects
// are screen pixels; a blit offsets them into the buffer by the scroll.
//
// The slices swap backing arrays in frameCommit, so the steady state
// allocates nothing.
type frameState struct {
	restore []image.Rectangle // chrome rects drawn last frame
	current []image.Rectangle // chrome rects collected this frame
	seen    bool              // a frameCommit has run
}

// frameBegin starts this frame's chrome collection.
func (s *shell) frameBegin() {
	s.frame.current = s.frame.current[:0]
}

// frameNote records one chrome rect actually painted this frame. An empty
// rect is ignored because nothing was painted there.
func (s *shell) frameNote(r image.Rectangle) {
	if r.Empty() {
		return
	}

	s.frame.current = append(s.frame.current, r)
}

// frameRestore returns the chrome rects of the last committed frame. The
// caller must only read the slice.
func (s *shell) frameRestore() []image.Rectangle {
	return s.frame.restore
}

// frameSeen reports whether a frame has been committed.
func (s *shell) frameSeen() bool {
	return s.frame.seen
}

// frameCommit makes this frame's rects the restore set and recycles the old
// restore slice as the empty current slice.
func (s *shell) frameCommit() {
	s.frame.restore, s.frame.current = s.frame.current, s.frame.restore[:0]
	s.frame.seen = true
}

// frameRect rounds a painted float rectangle outward to whole pixels, so a
// partially covered edge pixel is restored too.
func frameRect(x, y, w, h float64) image.Rectangle {
	return image.Rect(
		int(math.Floor(x)), int(math.Floor(y)),
		int(math.Ceil(x+w)), int(math.Ceil(y+h)),
	)
}
