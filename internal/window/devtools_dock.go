package window

const (
	// devDockWidth is the default width of the right-hand panel.
	devDockWidth = 340.0

	// devDockMin is the narrowest the panel may become while resizing.
	devDockMin = 220.0

	// devDockGrab is the half-width of the resize edge beside the panel.
	devDockGrab = 4.0

	// devDockGrabWidth spans the resize edge on both sides.
	devDockGrabWidth = 2 * devDockGrab
)

// devDockRect places the devtools panel on the right edge, full height. A
// width larger than the window is clamped, so a phone shows the panel alone.
func devDockRect(screenW, screenH int, width float64) devRect {
	if width < 0 {
		width = 0
	}

	if width > float64(screenW) {
		width = float64(screenW)
	}

	return devRect{X: float64(screenW) - width, Y: 0, W: width, H: float64(screenH)}
}

// devDockWidth returns the dock width for this screen: the stored drag
// width, else the default, never wider than the window and never under the
// minimum on a screen that can hold it.
func (s *shell) devDockWidth() float64 {
	w := s.dev.dockW
	if w <= 0 {
		w = devDockWidth
	}

	if max := float64(s.screenW); w > max {
		w = max
	}

	if w < devDockMin && float64(s.screenW) > devDockMin {
		w = devDockMin
	}

	return w
}

// devInRect reports whether a point is inside a rectangle.
func devInRect(r devRect, x, y float64) bool {
	return x >= r.X && x <= r.X+r.W && y >= r.Y && y <= r.Y+r.H
}

// devOnEdge reports the resize strip just left of the dock. It is off on a
// screen the dock already fills.
func devOnEdge(panel devRect, x, y int) bool {
	if panel.X <= devDockGrab {
		return false
	}

	edge := devRect{X: panel.X - devDockGrab, W: devDockGrabWidth, H: panel.H}

	return devInRect(edge, float64(x), float64(y))
}
