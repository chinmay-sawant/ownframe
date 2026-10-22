package a11y

import (
	"strings"

	"github.com/chinmay-sawant/blinkless/markup"
)

// NodeText is the flattened descendant text of n.
func NodeText(n *markup.Node) string {
	var out strings.Builder
	var walk func(*markup.Node)
	walk = func(node *markup.Node) {
		if node.Type == markup.TypeText {
			out.WriteString(node.Text)
			return
		}
		if node.Name == "script" || node.Name == "style" {
			return
		}
		for _, child := range node.Children {
			walk(child)
		}
	}
	walk(n)
	return strings.Join(strings.Fields(out.String()), " ")
}
