//go:build linux && !android && !ios && !js

package window

import (
	"github.com/chinmay-sawant/ownframe/internal/atspi"
)

// a11yHost holds the Linux AT-SPI publisher. When an AT-SPI bus is connected,
// sync publishes the accessible tree on redraw.
type a11yHost struct {
	pub       *atspi.Publisher
	gen       uint64
	published bool
}

func (s *shell) initA11y() {
	s.a11y.pub = atspi.NewPublisher(atspi.NewLiveBus())
}

func (s *shell) closeA11y() {
	if s.a11y.pub != nil {
		_ = s.a11y.pub.Close()
	}
}

func (s *shell) syncA11y() {
	if s.a11y.pub == nil || !s.a11y.pub.Ready() {
		return
	}
	x, y := 0, 0
	if s.a11yOrigin != nil {
		x, y = s.a11yOrigin()
	}
	s.a11y.pub.SetOrigin(x, y)
	s.a11y.pub.SetSize(s.screenW, s.screenH)
	s.publishA11y()
}
