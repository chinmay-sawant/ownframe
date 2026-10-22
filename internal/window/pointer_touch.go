package window

import (
	"github.com/hajimehoshi/ebiten/v2"
)

// touch tracks live fingers, scrolling, capture, and pinch.
// A system-generated mouse click suppresses the duplicate touch tap.
func (s *shell) touch(clicked bool, frameW, frameH int) error {
	ids := ebiten.AppendTouchIDs(nil)
	now := make([]touchPos, 0, len(ids))

	for _, id := range ids {
		x, y := ebiten.TouchPosition(id)
		now = append(now, touchPos{id: id, x: x, y: y})
	}

	return s.applyTouches(now, clicked, frameW, frameH)
}

// freshTouches returns the ids in now that are not in prev.
func freshTouches(now, prev []ebiten.TouchID) []ebiten.TouchID {
	var fresh []ebiten.TouchID

	for _, id := range now {
		if !hasTouch(prev, id) {
			fresh = append(fresh, id)
		}
	}

	return fresh
}

func hasTouch(ids []ebiten.TouchID, want ebiten.TouchID) bool {
	for _, id := range ids {
		if id == want {
			return true
		}
	}

	return false
}
