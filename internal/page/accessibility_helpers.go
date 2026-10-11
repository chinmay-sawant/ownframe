package page

import (
	"github.com/chinmay-sawant/blinkless/markup"
	"github.com/chinmay-sawant/ownframe/internal/host"
	"strings"
)

func hiddenAccessibilityNode(n *markup.Node) bool {
	if hasAttr(n, "hidden") || strings.EqualFold(n.Attrs["aria-hidden"], "true") {
		return true
	}
	switch n.Name {
	case "head", "script", "style", "template":
		return true
	}
	return false
}

func hasAttr(n *markup.Node, name string) bool { _, ok := n.Attrs[name]; return ok }

func canActivateAccessibility(c Control) bool {
	return c.Tag == "button" || c.Tag == "select" || c.Type == "checkbox" || c.Type == "radio" || c.Type == "file" || c.Type == "button" || c.Type == "submit" || c.Type == "reset"
}

func textNamedByParent(role host.AccessibilityRole) bool {
	switch role {
	case host.AccessibilityButton, host.AccessibilityLink, host.AccessibilityHeading, host.AccessibilityParagraph, host.AccessibilityListItem, host.AccessibilityTextbox, host.AccessibilityCheckbox, host.AccessibilityRadio, host.AccessibilityCombobox:
		return true
	default:
		return false
	}
}

func accessibilityID(n *markup.Node, path string, ids map[string][]*markup.Node) host.AccessibilityNodeID {
	if id := n.Attrs["id"]; id != "" && len(ids[id]) == 1 {
		return host.AccessibilityNodeID("id:" + id)
	}
	return host.AccessibilityNodeID("path:" + path)
}
