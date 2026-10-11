//go:build android || ios

package window

import (
	"image"

	"github.com/chinmay-sawant/ownframe/internal/host"
)

// imeResize restarts the session when the caret box moved, because Ebiten
// freezes the caret bounds a session reports at its start. A keyboard
// opening lifts the composer, and the restart makes Ebiten recompute the
// virtual keyboard shift against the new position instead of panning the
// canvas from the old one.
func (s *shell) imeResize() {
	target, ok := s.app.(host.IME)
	if !ok || s.ime.field == "" {
		return
	}

	box, _, _, _, ok := target.IMEContext()
	if !ok || imeNear(s.imeScreenRect(box), s.ime.caret) {
		return
	}

	s.ime.composer.Confirm()
	s.ime.last = ""
}

// imeNear reports whether two caret boxes are close enough that the
// rasterizer's rounding should not restart the session. Only a real layout
// move, such as the keyboard opening, is worth a restart.
func imeNear(a, b image.Rectangle) bool {
	const slack = 4

	return abs(a.Min.X-b.Min.X) <= slack && abs(a.Min.Y-b.Min.Y) <= slack &&
		abs(a.Max.X-b.Max.X) <= slack && abs(a.Max.Y-b.Max.Y) <= slack
}

func abs(n int) int {
	if n < 0 {
		return -n
	}

	return n
}
