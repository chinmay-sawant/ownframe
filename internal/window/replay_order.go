package window

import "github.com/chinmay-sawant/blinkless/layout"

type orderCache struct {
	display *layout.Display
	safe    bool
}

// A scroll buffer can precede viewport layers only if paint order already
// puts those layers last. Otherwise draw directly in engine order.
func (s *shell) bufferOrderSafe() bool {
	if s.orderCache.display == s.display {
		return s.orderCache.safe
	}
	s.orderCache = orderCache{display: s.display}
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
		if !viewport && viewportSeen {
			return false
		}
		viewportSeen = viewportSeen || viewport
	}
	s.orderCache.safe = true
	return true
}
