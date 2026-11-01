package window

import "github.com/chinmay-sawant/ownframe/internal/host"

func (s *shell) layoutSize(w, h int) (int, int) {
	if _, ok := s.app.(host.ViewportSizer); ok && !s.viewLocked() {
		return max(1, w), max(1, h)
	}
	return s.app.Clamp(w, h)
}

func (s *shell) setLayoutSize(w, h int) {
	if viewport, ok := s.app.(host.ViewportSizer); ok {
		viewport.SetViewport(w, h)
		return
	}
	s.app.SetSize(w, h)
}
