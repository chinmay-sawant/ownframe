package ownframe

// Element returns a retained handle for id.
func (t Testing) Element(id string) *Element {
	p := innerPage(t.page)
	if p == nil {
		return nil
	}

	return p.Element(id)
}

// Elements returns all retained handles.
func (t Testing) Elements() []*Element {
	p := innerPage(t.page)
	if p == nil {
		return nil
	}

	return p.Elements()
}

// ElementRegistry exposes the retained element registry.
func (t Testing) ElementRegistry() *ElementRegistry {
	p := innerPage(t.page)
	if p == nil {
		return nil
	}

	return p.ElementRegistry()
}

// Invalidate marks the region the element id changed.
func (t Testing) Invalidate(id string) {
	p := innerPage(t.page)
	if p == nil {
		return
	}
	p.Invalidate(id)
}
