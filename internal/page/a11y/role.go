package a11y

import (
	"github.com/chinmay-sawant/blinkless/markup"
	"github.com/chinmay-sawant/ownframe/internal/host"
)

// Role is the implicit or ARIA role of n.
func Role(n *markup.Node) (host.AccessibilityRole, bool) {
	if role, ok := SupportedRole(n.Attrs["role"]); ok {
		return role, true
	}
	switch n.Name {
	case "main":
		return host.AccessibilityMain, true
	case "nav":
		return host.AccessibilityNavigation, true
	case "section":
		if n.Attrs["aria-label"] != "" || n.Attrs["aria-labelledby"] != "" {
			return host.AccessibilityGroup, true
		}
	case "form", "fieldset":
		return host.AccessibilityGroup, true
	case "h1", "h2", "h3", "h4", "h5", "h6":
		return host.AccessibilityHeading, true
	case "p":
		return host.AccessibilityParagraph, true
	case "ul", "ol":
		return host.AccessibilityList, true
	case "li":
		return host.AccessibilityListItem, true
	case "a":
		if n.Attrs["href"] != "" {
			return host.AccessibilityLink, true
		}
	case "img":
		if n.Attrs["alt"] != "" || n.Attrs["title"] != "" || n.Attrs["aria-label"] != "" || n.Attrs["aria-labelledby"] != "" {
			return host.AccessibilityImage, true
		}
	case "button":
		return host.AccessibilityButton, true
	case "textarea":
		return host.AccessibilityTextbox, true
	case "select":
		return host.AccessibilityCombobox, true
	case "input":
		return InputRole(n.Attrs["type"])
	}
	return "", false
}
