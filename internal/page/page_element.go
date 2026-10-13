package page

import (
	"github.com/chinmay-sawant/ownframe/internal/element"
)

func (p *Page) ensureElements() *element.Registry {
	if p == nil {
		return element.NewRegistry()
	}
	if p.elements == nil {
		p.elements = element.NewRegistry()
	}
	return p.elements
}

// Element returns a retained handle for id.
func (p *Page) Element(id string) *element.Node {
	return p.RetainedElement(id)
}

// ElementHandle is an alias for Element.
func (p *Page) ElementHandle(id string) *element.Node {
	return p.RetainedElement(id)
}

// RetainedElement returns or creates a retained node handle.
func (p *Page) RetainedElement(id string) *element.Node {
	r := p.ensureElements()
	n, ok := r.Get(id)
	if ok && n != nil {
		return n
	}
	n = element.NewNode(id, "")
	r.Put(n)
	return n
}

// RetainedNode is an alias for RetainedElement.
func (p *Page) RetainedNode(id string) *element.Node {
	return p.RetainedElement(id)
}

// Handle is an alias for RetainedElement.
func (p *Page) HandleElement(id string) *element.Node {
	return p.RetainedElement(id)
}

// Elements returns all retained handles.
func (p *Page) Elements() []*element.Node {
	return p.ensureElements().All()
}

// RetainedRoot returns the retained tree root when set.
func (p *Page) RetainedRoot() *element.Node {
	r := p.ensureElements()
	all := r.All()
	if len(all) == 0 {
		return nil
	}
	return all[0]
}

// ElementRegistry exposes the underlying registry.
func (p *Page) ElementRegistry() *element.Registry {
	return p.ensureElements()
}
