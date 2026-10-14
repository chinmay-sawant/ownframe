// Package element retains element identity between parses.
//
// ownframe renders through a display list today. This package adds a small
// retained node tree keyed by IDs so mount/update/unmount and event
// delegation can attach to a stable instance instead of box hits.
package element

// Kind tags a node (button, input, ...).
type Kind string

// Layout keeps CSS layout geometry separate from metadata and paint.
type Layout struct {
	X, Y, W, H float64
}

// Node is a retained element instance.
type Node struct {
	ID        string
	Key       string
	Kind      Kind
	Attrs     map[string]string
	Metadata  map[string]any
	Children  []*Node
	Parent    *Node
	Handlers  map[string]Handler
	Mounted   bool
	Updated   bool
	Removed   bool
	Lifecycle Lifecycle
	Disposed  bool
	Layout    Layout
	Painted   bool
}

// NewNode creates an empty retained node.
func NewNode(id string, kind Kind) *Node {
	return &Node{
		ID:       id,
		Kind:     kind,
		Attrs:    map[string]string{},
		Metadata: map[string]any{},
		Handlers: map[string]Handler{},
	}
}

// Append adds a child and fixes parent pointers.
func (n *Node) Append(c *Node) {
	if n == nil || c == nil {
		return
	}
	c.Parent = n
	n.Children = append(n.Children, c)
}

// Remove detaches child if present.
func (n *Node) Remove(c *Node) bool {
	if n == nil || c == nil {
		return false
	}
	for i, ch := range n.Children {
		if ch == c {
			n.Children = append(n.Children[:i], n.Children[i+1:]...)
			c.Parent = nil
			c.Removed = true
			return true
		}
	}
	return false
}
