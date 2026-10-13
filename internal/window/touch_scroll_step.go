package window

import (
	"time"

	"github.com/chinmay-sawant/ownframe/internal/host"
)

func (s *shell) stepTouchScroll(now time.Time) {
	if s.scrollMotion.dragging || s.scrollMotion.vx == 0 && s.scrollMotion.vy == 0 {
		return
	}
	s.configureTouchScroll()
	if s.stretched() || !s.allowPageScroll() {
		s.scrollMotion.cancel()
		return
	}
	dx, dy := s.scrollMotion.step(now)
	if dx == 0 && dy == 0 {
		return
	}
	contentW, contentH := s.contentSize()
	viewW, viewH := s.viewSize()
	x, y := clampScroll(s.scrollX+dx, s.scrollY+dy, contentW, contentH, viewW, viewH)
	if x != s.scrollX+dx || y != s.scrollY+dy {
		s.scrollMotion.cancel()
	}
	s.scrollX, s.scrollY = x, y
}

func (s *shell) configureTouchScroll() {
	if config, ok := s.app.(host.TouchScrollConfig); ok {
		s.scrollMotion.configure(config.TouchScrollSensitivity(), config.TouchScrollDeceleration())
	}
}
