package page

import (
	"strings"

	"github.com/chinmay-sawant/blinkless/markup"
	"github.com/chinmay-sawant/ownframe/internal/host"
)

func (b *accessibilityBuilder) makeAccessibilityNode(n *markup.Node, id, parent host.AccessibilityNodeID, role host.AccessibilityRole) host.AccessibilityNode {
	node := host.AccessibilityNode{ID: id, Parent: parent, ElementID: n.Attrs["id"], Role: role}
	node.Name = b.accessibilityName(n, role)
	node.Description = b.accessibilityDescription(n)
	ariaDisabled := strings.EqualFold(n.Attrs["aria-disabled"], "true")
	node.Disabled = ariaDisabled
	if hasAttr(n, "disabled") {
		node.Disabled = true
	}

	if node.ElementID == "" || len(b.ids[node.ElementID]) != 1 {
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
	c, ok := b.page.control(node.ElementID)
	if !ok {
		return node
	}
	node.Value = c.Value
	node.Disabled = node.Disabled || c.Disabled
	node.Focusable = !c.Disabled
	node.Focused = b.page.FocusID() == c.ID
	if node.Focused {
		b.focused = id
	}
	node.Password = strings.EqualFold(c.Type, "password")
	if node.Password {
		node.Value = ""
	}
	if c.Type == "checkbox" || c.Type == "radio" {
		node.Checked, node.HasChecked = c.Checked, true
	}
	if c.Disabled {
		return node
	}
	node.Actions = append(node.Actions, host.AccessibilityFocus)
	if node.Disabled {
		return node
	}
	if canActivateAccessibility(c) {
		node.Actions = append(node.Actions, host.AccessibilityActivate)
	}
	if canEdit(c) && c.Type != "file" && !hasAttr(n, "readonly") && !strings.EqualFold(n.Attrs["aria-readonly"], "true") && !ariaDisabled {
		node.Actions = append(node.Actions, host.AccessibilitySetValue)
	}
	return node
}
