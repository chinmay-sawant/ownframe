package a11y

import (
	"strings"

	"github.com/chinmay-sawant/blinkless/markup"
	"github.com/chinmay-sawant/ownframe/internal/host"
)

func (b *builder) makeNode(
	n *markup.Node,
	id, parent host.AccessibilityNodeID,
	role host.AccessibilityRole,
) host.AccessibilityNode {
	node := host.AccessibilityNode{
		ID:        id,
		Parent:    parent,
		ElementID: n.Attrs["id"],
		Role:      role,
	}
	node.Name = b.names.Name(n, role)
	node.Description = b.names.Description(n)
	applyMarkupControl(&node, n)
	ariaDisabled := strings.EqualFold(n.Attrs["aria-disabled"], "true")
	node.Disabled = ariaDisabled
	if HasAttr(n, "disabled") {
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

func (b *builder) applyControl(node host.AccessibilityNode, n *markup.Node, ariaDisabled bool) host.AccessibilityNode {
	c, ok := b.controls[node.ElementID]
	if !ok {
		return node
	}
	node.Value = c.Value
	node.Disabled = node.Disabled || c.Disabled
	node.Focusable = !c.Disabled
	node.Focused = b.focusID == c.ID
	if node.Focused {
		b.focused = node.ID
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
	if CanActivate(c.Tag, c.Type) {
		node.Actions = append(node.Actions, host.AccessibilityActivate)
	}
	readonly := HasAttr(n, "readonly") || strings.EqualFold(n.Attrs["aria-readonly"], "true")
	if c.Editable && c.Type != "file" && !node.Password && !readonly && !ariaDisabled {
		node.Actions = append(node.Actions, host.AccessibilitySetValue)
	}
	return node
}

// Anonymous controls have no page action target. Preserve their declared
// state without advertising input operations the page cannot route.
func applyMarkupControl(node *host.AccessibilityNode, n *markup.Node) {
	if n.Name != "input" && n.Name != "textarea" && n.Name != "select" {
		return
	}
	kind := strings.ToLower(n.Attrs["type"])
	node.Password = kind == "password"
	if !node.Password {
		node.Value = n.Attrs["value"]
		if n.Name == "textarea" {
			node.Value = NodeText(n)
		}
	}
	if kind == "checkbox" || kind == "radio" {
		node.Checked, node.HasChecked = HasAttr(n, "checked"), true
	}
}
