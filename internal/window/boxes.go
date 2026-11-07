package window

import (
	"github.com/chinmay-sawant/blinkless/layout"
	"github.com/chinmay-sawant/ownframe/internal/host"
)

// boxes returns the active layout boxes without allocating a defensive copy.
// It prefers the current display's box list, then queries host.BoxSlicer if
// supported, and falls back to s.app.Boxes().
func (s *shell) boxes() []layout.Box {
	if s.display != nil && len(s.display.Boxes) > 0 {
		return s.display.Boxes
	}
	if slicer, ok := s.app.(host.BoxSlicer); ok {
		return slicer.BoxesSlice()
	}
	if s.app != nil {
		return s.app.Boxes()
	}

	return nil
}

// boxAt returns the innermost box containing (x, y), or (layout.Box{}, false).
func (s *shell) boxAt(x, y float64) (layout.Box, bool) {
	return devBoxAt(s.boxes(), x, y)
}
