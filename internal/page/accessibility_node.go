package page

import (
	"strings"

	"github.com/chinmay-sawant/blinkless/markup"
	"github.com/chinmay-sawant/ownframe/internal/host"
	"github.com/chinmay-sawant/ownframe/internal/page/a11y"
)

func (b *accessibilityBuilder) makeAccessibilityNode(n *markup.Node, id, parent host.AccessibilityNodeID, role host.AccessibilityRole) host.AccessibilityNode {
	node := host.AccessibilityNode{ID: id, Parent: parent, ElementID: n.Attrs["id"], Role: role}
	node.Name = b.names.Name(n, role)
	node.Description = b.names.Description(n)
	ariaDisabled := strings.EqualFold(n.Attrs["aria-disabled"], "true")
	node.Disabled = ariaDisabled
	if a11y.HasAttr(n, "disabled") {
		node.Disabled = true
	}

	if node.ElementID == "" || len(b.names.IDs[node.ElementID]) != 1 {
		if box, ok := b.nodeBox[n]; ok {
			node.Bounds = host.AccessibilityBounds{X: box.X, Y: box.Y, W: box.W, H: box.H}
			node.HasBounds = true
		}
		return node
	}
	if b.counts[node.ElementID] == 1 {
		box := b.boxes[node.ElementID]
		node.Bounds = host.AccessibilityBounds{X: box.X, Y: box.Y, W: box.W, H: box.H}
		node.HasBounds = true
	}
	return b.applyControl(node, n, ariaDisabled)
}
