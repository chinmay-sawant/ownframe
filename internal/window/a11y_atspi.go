//go:build linux && !android && !ios && !js

package window

import "github.com/chinmay-sawant/ownframe/internal/atspi"

// a11yHost holds the Linux AT-SPI publisher. When an AT-SPI bus is connected,
// sync publishes the accessible tree on redraw.
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
