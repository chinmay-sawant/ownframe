package element

// GetOrCreate reuses an instance when the key is stable.
func (r *Registry) GetOrCreate(key string, kind Kind) *Node {
	if r == nil {
		return NewNode(key, kind)
	}
	if n, ok := r.nodes[key]; ok && n != nil && !n.Disposed {
		n.Kind = kind
		return n
	}
	n := NewNode(key, kind)
	n.Key = key
	r.Put(n)
	return n
}

// Retain is an alias for GetOrCreate.
func (r *Registry) Retain(key string, kind Kind) *Node {
	return r.GetOrCreate(key, kind)
}

// Dispose removes one instance.
func (r *Registry) Dispose(key string) bool {
	if r == nil {
		return false
	}
	n, ok := r.nodes[key]
	if !ok || n == nil {
		return false
	}
	n.Dispose()
	delete(r.nodes, key)
	return true
}

// DisposeMissing removes instances not in alive.
func (r *Registry) DisposeMissing(alive map[string]bool) []string {
	if r == nil {
		return nil
	}
	var removed []string
	for key := range r.nodes {
		if alive != nil && alive[key] {
			continue
		}
		if r.Dispose(key) {
			removed = append(removed, key)
		}
	}
	return removed
}

// All returns instances in insertion order.
func (r *Registry) All() []*Node {
	if r == nil {
		return nil
	}
	out := make([]*Node, 0, len(r.order))
	for _, key := range r.order {
		if n, ok := r.nodes[key]; ok && n != nil {
			out = append(out, n)
		}
	}
	return out
}

// Len reports stored instances.
func (r *Registry) Len() int {
	if r == nil {
		return 0
	}
	return len(r.nodes)
}

// Clear removes every instance.
func (r *Registry) Clear() {
	if r == nil {
		return
	}
	for key := range r.nodes {
		if n := r.nodes[key]; n != nil {
			n.Dispose()
		}
	}
	r.nodes = map[string]*Node{}
	r.order = nil
}
