package window

import "github.com/hajimehoshi/ebiten/v2"

// stretched reports that the painted page is being scaled to the viewport.
// The scale is only for the gap while a window or dock drag waits on a
// relayout. Once the page has committed its size, a page that cannot fit the
// viewport (a min-size clamp) is scrolled instead, so a click stays on the
// pixel under the pointer.
func (s *shell) stretched() bool {
	viewW, viewH := s.viewSize()
	frameW, frameH := s.frameSize()
	if frameW == viewW && frameH == viewH {
		return false
	}

	wantW, wantH := s.app.Clamp(viewW, viewH)
	haveW, haveH := s.app.Size()
	if haveW == wantW && haveH == wantH {
		return false
	}

	return true
}

func (s *shell) wheel() {
	if s.devWheel() {
		return
	}

	wheelX, wheelY := ebiten.Wheel()

	if s.zoomWheel(wheelY) {
		return
	}

	if s.stretched() || !s.allowPageScroll() {
		return
	}

	contentW, contentH := s.contentSize()
	viewW, viewH := s.viewSize()
	s.scrollX, s.scrollY = panScroll(
		s.scrollX, s.scrollY, wheelX, wheelY, contentW, contentH, viewW, viewH,
	)
}

func (s *shell) frameSize() (int, int) {
	if s.img != nil {
		bounds := s.img.Bounds()

		return bounds.Dx(), bounds.Dy()
	}

	if s.display != nil {
		return s.display.Width, s.display.Height
	}

	return s.app.Size()
}
