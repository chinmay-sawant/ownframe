package window

// viewSize is the page viewport: the window minus the devtools dock when the
// overlay is open. The dock shrinks the page like a docked browser inspector
// instead of covering it, so a page relayouts to the strip that stays visible.
func (s *shell) viewSize() (int, int) {
	return s.viewOf(s.screenW, s.screenH)
}

// viewOf is viewSize for an outside size that Layout has not published yet.
func (s *shell) viewOf(w, h int) (int, int) {
	if s.dev.on {
		dock := int(s.devDockWidth())
		if dock > w {
			dock = w
		}

		w -= dock
	}

	if w < 1 {
		w = 1
	}

	if h < 1 {
		h = 1
	}

	return w, h
}
