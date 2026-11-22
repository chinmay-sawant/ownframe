//go:build (!linux || android || ios || js) && !(windows && (amd64 || arm64))

package window

// a11yHost is a compile stub for hosts without a native adapter. macOS VoiceOver and
// Android TalkBack are not shipped. This path never reports success.
type a11yHost struct{}

func (s *shell) initA11y() {}

func (s *shell) syncA11y() {}

func (s *shell) closeA11y() {}

func (s *shell) runA11yActions() {}
