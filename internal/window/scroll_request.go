package window

import (
	"github.com/chinmay-sawant/ownframe/internal/host"
)

// applyScrollRequest consumes a page scroll request and moves the offset.
// The shell keeps ownership of the offset; the request is clamped to the
// content and cleared either way.
func (s *shell) applyScrollRequest() {
	requester, ok := s.app.(host.ScrollRequester)
	if !ok {
		return
	}

	req, ok := requester.TakeScroll()
	if !ok {
		return
	}
	if !s.allowPageScroll() {
		s.scrollX, s.scrollY = 0, 0
		return
	}

	contentW, contentH := s.contentSize()
	viewW, viewH := s.viewSize()
	if req.Absolute {
		s.scrollX, s.scrollY = clampScroll(req.X, req.Y, contentW, contentH, viewW, viewH)

		return
	}

	s.scrollX, s.scrollY = clampScroll(
		s.scrollX+req.X, s.scrollY+req.Y, contentW, contentH, viewW, viewH,
	)
}
