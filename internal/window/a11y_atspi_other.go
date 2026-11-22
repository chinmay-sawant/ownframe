//go:build !linux || ios || js

package window

// a11yHost is a compile stub. Windows UIA and macOS VoiceOver are not
// shipped. Android has its own host in a11y_android.go. This path never
// reports success.
type a11yHost struct{}

func (s *shell) initA11y() {}

func (s *shell) syncA11y() {}

func (s *shell) closeA11y() {}

func (s *shell) runA11yActions() {}
