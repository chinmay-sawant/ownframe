package window

import "github.com/chinmay-sawant/blinkless/layout"

// viewportOperations identifies paint that follows the viewport rather than
// extending the scrollable document.
func (s *shell) viewportOperations() map[uint64]bool {
	if s.display == nil {
		return nil
	}
	pin := s.viewportPinZ()
	ids := make(map[uint64]bool)
	for _, op := range s.display.Ops {
		if op.Fixed || pin > 0 && op.ZIndexSet && op.ZIndex >= pin {
			ids[op.ID] = true
		}
	}
	return ids
}

func viewportBox(b layout.Box, ids map[uint64]bool) bool {
	if len(b.OperationIDs) == 0 {
		return false
	}
	for _, id := range b.OperationIDs {
		if !ids[id] {
			return false
		}
	}
	return true
}
