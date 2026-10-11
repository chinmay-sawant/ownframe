//go:build !js

package window

func (s *shell) imeCancelSession() error {
	s.ime.err = nil
	s.ime.composer.Cancel()
	s.ime.field, s.ime.last = "", ""
	return s.ime.err
}

func (s *shell) imeShutdown() { _ = s.imeCancelSession() }
