//go:build js

package window

// imeState is empty in the browser build, which uses browser text input.
type imeState struct{}

// imeInit leaves platform text input to the browser.
func (s *shell) imeInit() {}

// imeUpdate leaves platform text input to the browser.
func (s *shell) imeUpdate() error { return nil }

func (s *shell) imeLayoutChanged(int, int) {}

// imeHandled reports that the platform did not consume input.
func (s *shell) imeHandled() bool { return false }

func (s *shell) imeCancelSession() error { return nil }

func (s *shell) imeShutdown() {}
