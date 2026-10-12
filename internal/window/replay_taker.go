package window

import "image"

// dirtyTaker is the page side of the partial replay. TakeDirtyRects returns
// the regions that changed since the previous call, in CSS pixels, and
// clears them.
type dirtyTaker interface {
	TakeDirtyRects() ([]image.Rectangle, bool)
}

// singleTaker is a page from before the damage list: one dirty rect.
type singleTaker interface {
	TakeDirty() (image.Rectangle, bool)
}

// listTaker adapts a singleTaker to the list method. Its one rect becomes a
// one-element list.
type listTaker struct{ singleTaker }

func (t listTaker) TakeDirtyRects() ([]image.Rectangle, bool) {
	rect, ok := t.TakeDirty()
	if !ok {
		return nil, false
	}

	return []image.Rectangle{rect}, true
}

// damageTaker resolves the page's damage source: the list method when the
// page has it, otherwise a wrapped single rect.
func (s *shell) damageTaker() (dirtyTaker, bool) {
	if taker, ok := s.app.(dirtyTaker); ok {
		return taker, true
	}

	if one, ok := s.app.(singleTaker); ok {
		return listTaker{one}, true
	}

	return nil, false
}
