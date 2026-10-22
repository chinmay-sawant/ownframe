//go:build !linux || android || ios || js

package window

// a11yHost is a compile stub. Windows UIA, macOS VoiceOver, and
// Android TalkBack are not shipped. This path never reports success.
type a11yHost struct{}

func (s *shell) initA11y() {}

func (s *shell) syncA11y() {}
