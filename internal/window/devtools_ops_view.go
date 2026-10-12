package window

// devOpsView is the page area an outline may land in: the window minus the
// dock, so a stroke under the panel or off the window is skipped. The dock
// covering the whole window leaves it empty.
func (s *shell) devOpsView() devRect {
	if s.dev.panel.W <= 0 {
		return devRect{W: float64(s.screenW), H: float64(s.screenH)}
	}

	return devRect{W: s.dev.panel.X, H: float64(s.screenH)}
}

// devRectsMeet reports whether two screen rects overlap.
func devRectsMeet(a, b devRect) bool {
	return a.X < b.X+b.W && b.X < a.X+a.W && a.Y < b.Y+b.H && b.Y < a.Y+a.H
}
