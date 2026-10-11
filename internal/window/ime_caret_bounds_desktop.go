//go:build !android && !ios && !js

package window

import (
	"image"

	"github.com/chinmay-sawant/blinkless/layout"
	"github.com/chinmay-sawant/ownframe/internal/host"
)

func (s *shell) imeCaretScreenRect(box layout.Box) image.Rectangle {
	if target, ok := s.app.(host.IMECaretAnchor); ok {
		if anchor, ok := target.IMEAnchor(); ok {
			box = anchor
		}
	}
	return s.imeScreenRect(box)
}
