package page

import (
	"strings"

	"github.com/chinmay-sawant/blinkless/markup"
	"github.com/chinmay-sawant/ownframe/internal/host"
)

func accessibilityRole(n *markup.Node) (host.AccessibilityRole, bool) {
	if role, ok := supportedAccessibilityRole(n.Attrs["role"]); ok {
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
		return inputAccessibilityRole(n.Attrs["type"])
	}
	return "", false
}

func inputAccessibilityRole(kind string) (host.AccessibilityRole, bool) {
	switch strings.ToLower(kind) {
	case "hidden":
		return "", false
	case "checkbox":
		return host.AccessibilityCheckbox, true
	case "radio":
		return host.AccessibilityRadio, true
	case "button", "submit", "reset", "file", "image":
		return host.AccessibilityButton, true
	case "range", "color":
		return "", false
	default:
		return host.AccessibilityTextbox, true
	}
}
