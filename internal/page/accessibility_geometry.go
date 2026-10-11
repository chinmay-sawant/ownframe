package page

import (
	"strings"

	"github.com/chinmay-sawant/blinkless/markup"
)

type accessibilityBoxKey struct{ tag, text string }

func accessibilityNodeBoxes(root *markup.Node, boxes []Box) map[*markup.Node]Box {
	nodes := map[accessibilityBoxKey][]*markup.Node{}
	var walk func(*markup.Node)
	walk = func(n *markup.Node) {
		if n == nil {
			return
		}
		if n.Type == markup.TypeElement && n.Attrs["id"] == "" {
			key := accessibilityBoxKey{tag: n.Name, text: nodeText(n)}
			nodes[key] = append(nodes[key], n)
		}
		for _, child := range n.Children {
			walk(child)
		}
	}
	walk(root)

	placed := map[accessibilityBoxKey][]Box{}
	for _, box := range boxes {
		if box.ID == "" {
			text := strings.Join(strings.Fields(box.Text), " ")
			key := accessibilityBoxKey{tag: box.Tag, text: text}
			placed[key] = append(placed[key], box)
		}
	}

	out := map[*markup.Node]Box{}
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
