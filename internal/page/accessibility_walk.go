package page

import (
	"strconv"
	"strings"

	"github.com/chinmay-sawant/blinkless/markup"
	"github.com/chinmay-sawant/ownframe/internal/host"
)

func (b *accessibilityBuilder) walk(n *markup.Node, parent host.AccessibilityNodeID, path string) {
	if n == nil {
		return
	}
	if n.Type == markup.TypeText {
		b.appendAccessibilityText(n, parent, path)
		return
	}
	if n.Type != markup.TypeElement || hiddenAccessibilityNode(n) {
		return
	}
	role, ok := accessibilityRole(n)
	if ok {
		id := accessibilityID(n, path, b.ids)
		node := b.makeAccessibilityNode(n, id, parent, role)
		b.tree.Nodes = append(b.tree.Nodes, node)
		b.nodeAt[id] = len(b.tree.Nodes) - 1
		parentIndex := b.nodeAt[parent]
		b.tree.Nodes[parentIndex].Children = append(b.tree.Nodes[parentIndex].Children, id)
		if node.Focused {
			b.focused = id
		}
		parent = id
	}
	for i, child := range n.Children {
		b.walk(child, parent, childPath(path, i))
	}
}

func (b *accessibilityBuilder) appendAccessibilityText(n *markup.Node, parent host.AccessibilityNodeID, path string) {
	name := strings.Join(strings.Fields(n.Text), " ")
	if name == "" || textNamedByParent(b.tree.Nodes[b.nodeAt[parent]].Role) {
		return
	}
	id := host.AccessibilityNodeID("path:" + path)
	node := host.AccessibilityNode{ID: id, Parent: parent, Role: host.AccessibilityText, Name: name}
	b.nodeAt[id] = len(b.tree.Nodes)
	b.tree.Nodes = append(b.tree.Nodes, node)
	i := b.nodeAt[parent]
	b.tree.Nodes[i].Children = append(b.tree.Nodes[i].Children, id)
}

func childPath(parent string, at int) string {
	part := strconv.Itoa(at)
	if parent == "" {
		return part
	}
	return parent + "." + part
}
