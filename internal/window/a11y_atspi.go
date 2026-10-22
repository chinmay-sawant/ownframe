//go:build linux && !android && !ios && !js

package window

import "github.com/chinmay-sawant/ownframe/internal/atspi"

// a11yHost holds the Linux AT-SPI publisher. The live D-Bus client is
// not implemented, so Live stays false and sync is a no-op until then.
type a11yHost struct {
	pub *atspi.Publisher
}

func (s *shell) initA11y() {
	s.a11y.pub = atspi.NewPublisher(atspi.NewLiveBus())
}

func (s *shell) syncA11y() {
	if s.a11y.pub == nil || !s.a11y.pub.Live() {
		return
	}
	s.publishA11y()
}
