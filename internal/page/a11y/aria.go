package a11y

import (
	"strings"

	"github.com/chinmay-sawant/ownframe/internal/host"
)

// SupportedRole maps a raw ARIA role token to a host role.
func SupportedRole(raw string) (host.AccessibilityRole, bool) {
	fields := strings.Fields(raw)
	if len(fields) == 0 {
		return "", false
	}
	switch strings.ToLower(fields[0]) {
	case "main":
		return host.AccessibilityMain, true
	case "navigation":
		return host.AccessibilityNavigation, true
	case "group", "region":
		return host.AccessibilityGroup, true
	case "heading":
		return host.AccessibilityHeading, true
	case "paragraph":
		return host.AccessibilityParagraph, true
	case "img":
		return host.AccessibilityImage, true
	case "link":
		return host.AccessibilityLink, true
	case "button":
		return host.AccessibilityButton, true
	case "textbox":
		return host.AccessibilityTextbox, true
	case "checkbox":
		return host.AccessibilityCheckbox, true
	case "radio":
		return host.AccessibilityRadio, true
	case "combobox":
		return host.AccessibilityCombobox, true
	case "list":
		return host.AccessibilityList, true
	case "listitem":
		return host.AccessibilityListItem, true
	default:
		return "", false
	}
}
