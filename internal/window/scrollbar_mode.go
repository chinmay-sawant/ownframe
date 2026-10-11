package window

import (
	"runtime"

	"github.com/chinmay-sawant/ownframe/internal/host"
)

func (s *shell) scrollbarsEnabled() bool {
	if s.stretched() || !s.allowPageScroll() {
		return false
	}

	mode := host.ScrollbarAuto
	if provider, ok := s.app.(host.ScrollbarModeProvider); ok {
		mode = provider.ScrollbarMode()
	}

	return scrollbarModeEnabled(mode, runtime.GOOS)
}

func scrollbarModeEnabled(mode host.ScrollbarMode, goos string) bool {
	switch mode {
	case host.ScrollbarVisible:
		return true
	case host.ScrollbarHidden:
		return false
	default:
		return goos != "android" && goos != "ios"
	}
}
