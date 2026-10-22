package page

import (
	"github.com/chinmay-sawant/blinkless/markup"
	"github.com/chinmay-sawant/ownframe/internal/host"
	"github.com/chinmay-sawant/ownframe/internal/page/a11y"
	"strings"
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

type accessibilityState struct {
	source   string
	root     *markup.Node
	document uint64
	retire   bool
}

func (p *Page) commitAccessibility(source string) {
	if p.a11yState() == nil {
		p.setA11yState(&accessibilityState{})
	}
	s := p.a11yState()
	if s.document == 0 || s.retire {
		s.document++
		s.retire = false
	}
	if s.source != source {
		s.root = nil
	}
	s.source = source
}

func (p *Page) invalidateAccessibilityDocument() {
	if p.a11yState() != nil {
		p.a11yState().retire = true
	}
}
