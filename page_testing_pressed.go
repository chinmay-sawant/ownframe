package ownframe

// PressedID is the control held by the current pointer gesture.
func (t Testing) PressedID() string {
	p := innerPage(t.page)
	if p == nil {
		return ""
	}

	return p.PressedID()
}

// Generation increases by one on every successful Redraw.
func (t Testing) Generation() uint64 {
	p := innerPage(t.page)
	if p == nil {
		return 0
	}

	return p.Generation()
}
