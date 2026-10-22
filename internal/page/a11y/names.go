package a11y

import "github.com/chinmay-sawant/blinkless/markup"

// Names holds the id, label, and parent indexes used to compute
// accessible names.
type Names struct {
	IDs     map[string][]*markup.Node
	Labels  map[string][]*markup.Node
	Parents map[*markup.Node]*markup.Node
}

// NewNames starts empty indexes.
func NewNames() Names {
	return Names{
		IDs:     map[string][]*markup.Node{},
		Labels:  map[string][]*markup.Node{},
		Parents: map[*markup.Node]*markup.Node{},
	}
}

// Index records node and its descendants.
func (n *Names) Index(node, parent *markup.Node) {
	if node == nil {
		return
	}
	n.Parents[node] = parent
	if id := node.Attrs["id"]; id != "" {
		n.IDs[id] = append(n.IDs[id], node)
	}
	if node.Name == "label" && node.Attrs["for"] != "" {
		n.Labels[node.Attrs["for"]] = append(n.Labels[node.Attrs["for"]], node)
	}
	for _, child := range node.Children {
		n.Index(child, node)
	}
}
