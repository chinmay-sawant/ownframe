package patch

import "github.com/chinmay-sawant/ownframe/internal/element"

// Ensure returns r, or a new registry when r is nil.
func Ensure(r *element.Registry) *element.Registry {
	if r == nil {
		return element.NewRegistry()
	}

	return r
}

// Node returns the retained handle for id, creating one when the registry
// does not already hold it.
func Node(r *element.Registry, id string) *element.Node {
	n, ok := r.Get(id)
	if ok && n != nil {
		return n
	}

	n = element.NewNode(id, "")
	r.Put(n)

	return n
}

// All returns the retained handles in insertion order.
func All(r *element.Registry) []*element.Node {
	return r.All()
}

// Root returns the first retained handle, or nil when the registry is empty.
func Root(r *element.Registry) *element.Node {
	all := r.All()
	if len(all) == 0 {
		return nil
	}

	return all[0]
}
