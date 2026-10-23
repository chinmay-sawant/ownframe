package a11y

import (
	"strings"

	"github.com/chinmay-sawant/blinkless/markup"
)

// LabelText is the label associated with n by for= or wrapping.
func (b Names) LabelText(n *markup.Node) string {
	id := n.Attrs["id"]
	for _, label := range b.Labels[id] {
		if v := NodeText(label); v != "" {
			return v
		}
	}
	for parent := b.Parents[n]; parent != nil; parent = b.Parents[parent] {
		if parent.Name == "label" {
			return NodeText(parent)
		}
	}
	return ""
}

// ReferencedText joins the unique nodes named by raw id tokens.
func (b Names) ReferencedText(raw string) string {
	var parts []string
	for _, id := range strings.Fields(raw) {
		if nodes := b.IDs[id]; len(nodes) == 1 {
			if v := NodeText(nodes[0]); v != "" {
				parts = append(parts, v)
			}
		}
	}
	return strings.Join(parts, " ")
}
