package window

import "github.com/chinmay-sawant/ownframe/internal/host"

// roles narrows shell.app to what each call site needs. Entry points
// keep host.Screen for compat; pointer, keyboard, and clipboard code
// routes through one role so fakes implement roles, not the surface.
func (s *shell) ptr() host.Pointer    { return s.app }
func (s *shell) kbd() host.Keyboard   { return s.app }
func (s *shell) clip() host.Clipboard { return s.app }
