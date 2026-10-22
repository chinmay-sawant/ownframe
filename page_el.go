package ownframe

// Element returns a retained handle for id.
func (p *Page) Element(id string) *Element { return innerPage(p).Element(id) }

// Elements returns all retained handles.
func (p *Page) Elements() []*Element { return innerPage(p).Elements() }

// ElementRegistry exposes the retained element registry.
// Spotify player tests keep identity across redraws through it.
func (p *Page) ElementRegistry() *ElementRegistry {
	return innerPage(p).ElementRegistry()
}

// Invalidate marks the region the element id changed.
func (p *Page) Invalidate(id string) { innerPage(p).Invalidate(id) }
