package page

import (
	"strings"

	"github.com/chinmay-sawant/blinkless/markup"
	"github.com/chinmay-sawant/ownframe/internal/host"
	"github.com/chinmay-sawant/ownframe/internal/page/a11y"
)

func (b *accessibilityBuilder) applyControl(node host.AccessibilityNode, n *markup.Node, ariaDisabled bool) host.AccessibilityNode {
	c, ok := b.page.control(node.ElementID)
	if !ok {
		return node
	}
	node.Value = c.Value
	node.Disabled = node.Disabled || c.Disabled
	node.Focusable = !c.Disabled
	node.Focused = b.page.FocusID() == c.ID
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
	if a11y.CanActivate(c.Tag, c.Type) {
		node.Actions = append(node.Actions, host.AccessibilityActivate)
	}
	readonly := a11y.HasAttr(n, "readonly") || strings.EqualFold(n.Attrs["aria-readonly"], "true")
	if canEdit(c) && c.Type != "file" && !readonly && !ariaDisabled {
		node.Actions = append(node.Actions, host.AccessibilitySetValue)
	}
	return node
}
