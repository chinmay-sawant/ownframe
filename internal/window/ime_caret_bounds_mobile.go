//go:build android || ios

package window

import (
	"image"

	"github.com/chinmay-sawant/blinkless/layout"
)

func (s *shell) imeCaretScreenRect(box layout.Box) image.Rectangle {
	return s.imeScreenRect(box)
}
