package web

import "github.com/chinmay-sawant/ownframe/internal/host"

// roles narrows server.app to what each handler needs. Entry points
// keep host.Screen for compat; click and typing handlers route through
// one role so fakes implement roles, not the surface.
func (s *server) ptr() host.Pointer  { return s.app }
func (s *server) kbd() host.Keyboard { return s.app }
