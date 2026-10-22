package scroll

import "github.com/chinmay-sawant/ownframe/internal/host"

// State is the scroll state of a page. The window writes it and the
// page loop reads it.
type State struct {
	Req       host.Scroll
	Has       bool
	Windowing Windowing
}

// To queues an absolute scroll move.
func (s *State) To(x, y int) {
	s.Req = host.Scroll{X: x, Y: y, Absolute: true}
	s.Has = true
}

// By queues a scroll delta. A delta after To moves the pending target.
func (s *State) By(dx, dy int) {
	s.Req.X += dx
	s.Req.Y += dy
	s.Has = true
}

// Take returns the queued request and clears it.
func (s *State) Take() (host.Scroll, bool) {
	req, ok := s.Req, s.Has
	s.Req, s.Has = host.Scroll{}, false
	return req, ok
}
