package window

import "math"

const (
	// devDockWidth is the widest the right-hand panel opens, on a roomy
	// screen.
	devDockWidth = 340.0

	// devDockMin is the narrowest the panel may become while resizing.
	devDockMin = 220.0

	// devDockPageMin is the page strip the panel must leave beside it on a
	// desktop screen with room for that strip.
	devDockPageMin = 320.0

	// devDockRoomyW is the window width at which the panel opens at its full
	// devDockWidth. Narrower windows give the panel a share of the window
	// instead, so the page keeps a usable strip.
	devDockRoomyW = 1180

	// devDockShare is the fraction of a window narrower than devDockRoomyW
	// the panel takes, between devDockMin and devDockWidth.
	devDockShare = 0.28

	// devDockGrab is the half-width of the resize edge beside the panel.
	devDockGrab = 4.0

	// devDockGrabWidth spans the resize edge on both sides.
	devDockGrabWidth = 2 * devDockGrab
)

// devDockDefaultWidth is the panel width on a window with no stored drag
// width: devDockWidth on a roomy screen, else devDockShare of the window
// between devDockMin and devDockWidth, rounded down. It never exceeds the
// window, so a narrow screen still shows the panel alone.
func devDockDefaultWidth(screenW int) float64 {
	if screenW >= devDockRoomyW {
		return devDockWidth
	}

	w := float64(screenW) * devDockShare
	if w > devDockWidth {
		w = devDockWidth
	}

	if w < devDockMin {
		w = devDockMin
	}

	w = math.Floor(w)

	if maxW := float64(screenW); w > maxW {
		w = maxW
	}

	return w
}

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

// devDockWidth returns the dock width for this screen: the stored drag width
// clamped so the page keeps devDockPageMin beside it, else the responsive
// default. A screen narrower than devDockMin keeps the old clamp: the width
// is capped at the screen, so the default fills it.
func (s *shell) devDockWidth() float64 {
	screenW := float64(s.screenW)

	w := s.dev.dockW
	if w <= 0 {
		w = devDockDefaultWidth(s.screenW)
	}

	if screenW < devDockMin {
		if w > screenW {
			w = screenW
		}

		return w
	}

	if maxW := screenW - devDockPageMin; w > maxW {
		w = maxW
	}

	if w < devDockMin {
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

// devOwnsPoint reports whether the dock owns the pointer: inside the panel
// or on its resize edge. A custom-chrome draggable strip must not start a
// window drag there, or the header close key never receives the press.
func (s *shell) devOwnsPoint(x, y int) bool {
	if !s.dev.on {
		return false
	}

	panel := s.dev.panel
	if panel.W <= 0 || panel.H <= 0 {
		if s.dev.mobile {
			panel = devMobileRect(s.screenW, s.screenH, s.dev.mobileExpanded)
		} else {
			panel = devDockRect(s.screenW, s.screenH, s.devDockWidth())
		}
	}

	if devInRect(panel, float64(x), float64(y)) {
		return true
	}

	return !s.dev.mobile && devOnEdge(panel, x, y)
}
