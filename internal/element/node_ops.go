package element

// Dispose releases a node and its subtree.
func (n *Node) Dispose() {
	if n == nil || n.Disposed {
		return
	}
	children := append([]*Node(nil), n.Children...)
	n.Disposed = true
	n.Removed = true
	for _, c := range children {
		if c != nil {
			c.Dispose()
		}
	}
	Unmount(n)
	n.Children = nil
	n.Parent = nil
}

// Find returns a descendant by id.
func (n *Node) Find(id string) *Node {
	if n == nil {
		return nil
	}
	if n.ID == id {
		return n
	}
	for _, c := range n.Children {
		if f := c.Find(id); f != nil {
			return f
		}
	}
	return nil
}

// MetadataValue reads one metadata value.
func (n *Node) MetadataValue(key string) (any, bool) {
	if n == nil || n.Metadata == nil {
		return nil, false
	}
	v, ok := n.Metadata[key]
	return v, ok
}
