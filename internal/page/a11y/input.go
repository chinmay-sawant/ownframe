package a11y

import (
	"strings"

	"github.com/chinmay-sawant/ownframe/internal/host"
)

// InputRole is the role of an input with the given type.
func InputRole(kind string) (host.AccessibilityRole, bool) {
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
