package a11y

import (
	"strings"

	"github.com/chinmay-sawant/blinkless/layout"
	"github.com/chinmay-sawant/blinkless/markup"
)

type boxKey struct{ tag, text string }

// NodeBoxes maps anonymous nodes to boxes when the tag and text match
// unambiguously.
func NodeBoxes(root *markup.Node, boxes []layout.Box) map[*markup.Node]layout.Box {
	nodes := map[boxKey][]*markup.Node{}
	var walk func(*markup.Node)
	walk = func(n *markup.Node) {
		if n == nil {
			return
		}
		if n.Type == markup.TypeElement && n.Attrs["id"] == "" {
			key := boxKey{tag: n.Name, text: NodeText(n)}
			nodes[key] = append(nodes[key], n)
		}
		for _, child := range n.Children {
			walk(child)
		}
	}
	walk(root)

	placed := map[boxKey][]layout.Box{}
	for _, box := range boxes {
		if box.ID == "" {
			text := strings.Join(strings.Fields(box.Text), " ")
			key := boxKey{tag: box.Tag, text: text}
			placed[key] = append(placed[key], box)
		}
	}

	out := map[*markup.Node]layout.Box{}
	for key, group := range nodes {
		matches := placed[key]
		if len(group) != len(matches) {
			continue
		}
		for i, n := range group {
			out[n] = matches[i]
		}
	}
	return out
}
