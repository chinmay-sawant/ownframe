package a11y

import (
	"github.com/chinmay-sawant/blinkless/layout"
	"github.com/chinmay-sawant/blinkless/markup"
)

type boxKey struct{ tag, text string }

// NodeBoxes maps anonymous nodes to boxes when the tag and text match
// unambiguously.
func NodeBoxes(root *markup.Node, boxes []layout.Box) map[*markup.Node]layout.Box {
	return matchNodeBoxes(root, boxes, false)
}

// nodeBoxesForRoles maps only anonymous elements that carry an accessibility
// role, avoiding text flattening and box scanning for layout wrappers and
// thousands of syntax-highlighting spans that never enter the a11y tree.
func nodeBoxesForRoles(root *markup.Node, boxes []layout.Box) map[*markup.Node]layout.Box {
	return matchNodeBoxes(root, boxes, true)
}

func matchNodeBoxes(root *markup.Node, boxes []layout.Box, rolesOnly bool) map[*markup.Node]layout.Box {
	nodes := map[boxKey][]*markup.Node{}
	wantedTags := map[string]struct{}{}
	var walk func(*markup.Node)
	walk = func(n *markup.Node) {
		if n == nil {
			return
		}
		if n.Type == markup.TypeElement {
			if Hidden(n) {
				return
			}
			if n.Attrs["id"] == "" {
				include := true
				if rolesOnly {
					_, include = Role(n)
				}
				if include {
					key := boxKey{tag: n.Name, text: NodeText(n)}
					nodes[key] = append(nodes[key], n)
					wantedTags[n.Name] = struct{}{}
				}
			}
		}
		for _, child := range n.Children {
			walk(child)
		}
	}
	walk(root)
	if len(nodes) == 0 {
		return nil
	}

	placed := map[boxKey][]layout.Box{}
	for _, box := range boxes {
		if box.ID != "" || box.Tag == "" || box.Tag[0] == '#' {
			continue
		}
		if _, ok := wantedTags[box.Tag]; !ok {
			continue
		}
		key := boxKey{tag: box.Tag, text: collapseSpaces(box.Text)}
		if _, needed := nodes[key]; !needed {
			continue
		}
		placed[key] = append(placed[key], box)
	}

	out := make(map[*markup.Node]layout.Box, len(placed))
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
