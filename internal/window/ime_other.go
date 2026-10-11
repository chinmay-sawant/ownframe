//go:build !android && !ios

package window

// imeState is empty where the platform text input is not wired.
type imeState struct{}

// imeInit wires nothing.
func (s *shell) imeInit() {}

// imeUpdate does nothing.
func (s *shell) imeUpdate() error { return nil }

func (s *shell) imeLayoutChanged(int, int) {}

// imeHandled reports that the platform did not consume input.
func (s *shell) imeHandled() bool { return false }
