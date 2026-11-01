package window

import "github.com/hajimehoshi/ebiten/v2"

func setWindowRect(x, y, w, h int) {
	if setNativeWindowRect(x, y, w, h) {
		return
	}
	ebiten.SetWindowPosition(x, y)
	ebiten.SetWindowSize(w, h)
}

func windowPosition() (int, int) {
	if x, y, ok := nativeWindowPosition(); ok {
		return x, y
	}
	return ebiten.WindowPosition()
}

// Restore accuracy is independent of the drag loop tolerances.
func restoreRectMatches(actual, want workarea) bool { return actual == want }

func (s *shell) cursorPosition() (int, int) {
	if x, y, ok := s.nativeCursor.cursor(); ok {
		return x, y
	}
	return ebiten.CursorPosition()
}
