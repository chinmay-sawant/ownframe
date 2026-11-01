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
