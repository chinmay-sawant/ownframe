//go:build (!linux || ios || js) && !android && !(windows && (amd64 || arm64))

package window

// a11yHost is a compile stub for hosts without a native adapter. macOS
// VoiceOver is not shipped. Android has its own host in a11y_android.go and
// Windows (amd64, arm64) has a11y_uia_windows.go. This path never reports
// success.
type a11yHost struct{}

func (s *shell) initA11y() {}

func (s *shell) syncA11y() {}

func (s *shell) closeA11y() {}

func (s *shell) runA11yActions() {}
