package window

import (
	"math"

	"github.com/chinmay-sawant/ownframe/internal/host"
)

func (s *shell) viewLocked() bool {
	lock, ok := s.app.(host.ViewLock)

	return ok && lock.ViewLocked()
}

// holdView drops a pinch. The page stays at its own size.
func (s *shell) holdView() {
	s.fingers.zoom = 1
	s.fingers.span0 = 0
	s.pageZoom = 1
}

func (s *shell) touchZoomAllowed() bool {
	policy, ok := s.app.(host.TouchZoomPolicy)
	return !ok || policy.TouchZoomAllowed()
}

func (s *shell) holdTouchZoom() {
	s.fingers.zoom = 1
	s.fingers.span0 = 0
}

// fitBox scales content into the screen and centers it.
// The scale is the same on both axes, so a tall phone and a desktop window
// both show the whole page.
func fitBox(contentW, contentH, screenW, screenH int) (float64, float64, float64) {
	if contentW < 1 || contentH < 1 || screenW < 1 || screenH < 1 {
		return 1, 0, 0
	}

	scale := math.Min(
		float64(screenW)/float64(contentW),
		float64(screenH)/float64(contentH),
	)
	ox := (float64(screenW) - float64(contentW)*scale) / 2
	oy := (float64(screenH) - float64(contentH)*scale) / 2

	return scale, ox, oy
}

func (s *shell) fit() (float64, float64, float64) {
	if !s.viewLocked() {
		return 1, 0, 0
	}

	w, h := s.contentSize()
	viewW, viewH := s.viewSize()

	return fitBox(w, h, viewW, viewH)
}
