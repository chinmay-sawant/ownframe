package window

import "github.com/chinmay-sawant/blinkless/layout"

type orderCache struct {
	display  *layout.Display
	safe     bool
	viewport bool
}

// bufferOrderSafe reports that the scroll buffer can precede the viewport
// layers: paint order already puts those layers last.
func (s *shell) bufferOrderSafe() bool {
	s.scanOrder()

	return s.orderCache.safe
}

// hasViewportOps reports a page with fixed or viewport-pinned operations.
// Those layers are painted on the screen outside the buffer, so a partial
// blit would leave their old pixels behind; the frame repaints in full.
func (s *shell) hasViewportOps() bool {
	s.scanOrder()

	return s.orderCache.viewport
}

// scanOrder fills the order cache once per display. It records both facts in
// one pass so a partial frame pays for neither separately.
func (s *shell) scanOrder() {
	if s.orderCache.display == s.display {
		return
	}

	s.orderCache = orderCache{display: s.display, safe: true}
	viewportSeen := false

	for _, i := range s.display.Order {
		if i < 0 || i >= len(s.display.Ops) {
			continue
		}

		op := &s.display.Ops[i]
		if op.Kind == layout.DisplayOpNoop || op.Kind == layout.DisplayOpLinkURI {
			continue
		}

		viewport := op.Fixed || s.viewportPinZ() > 0 && op.ZIndexSet && op.ZIndex >= s.viewportPinZ()
		s.orderCache.viewport = s.orderCache.viewport || viewport

		if !viewport && viewportSeen {
			s.orderCache.safe = false
		}

		viewportSeen = viewportSeen || viewport
	}
}
