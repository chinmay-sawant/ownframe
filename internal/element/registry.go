package element

// Registry retains element instances by stable key across redraws.
type Registry struct {
	nodes map[string]*Node
	order []string
}

// NewRegistry returns an empty registry.
func NewRegistry() *Registry {
	return &Registry{nodes: map[string]*Node{}}
}

// KeyOf returns the stable key for a node.
func KeyOf(n *Node) string {
	if n == nil {
		return ""
	}
	if n.Key != "" {
		return n.Key
	}
	return n.ID
}

// Get returns an instance by key.
func (r *Registry) Get(key string) (*Node, bool) {
	if r == nil {
		return nil, false
	}
	n, ok := r.nodes[key]
	return n, ok
}

// Put records an instance.
func (r *Registry) Put(n *Node) {
	if r == nil || n == nil {
		return
	}
	key := KeyOf(n)
	if _, ok := r.nodes[key]; !ok {
		r.order = append(r.order, key)
	}
	r.nodes[key] = n
}
