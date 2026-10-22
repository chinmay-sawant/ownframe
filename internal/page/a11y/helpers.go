package a11y

import "github.com/chinmay-sawant/ownframe/internal/host"

// CanActivate reports that a control with tag and kind advertises Activate.
func CanActivate(tag, kind string) bool {
	return tag == "button" || tag == "select" || kind == "checkbox" || kind == "radio" || kind == "file" || kind == "button" || kind == "submit" || kind == "reset"
}

// TextNamedByParent reports that a parent with role already names its text.
func TextNamedByParent(role host.AccessibilityRole) bool {
	switch role {
	case host.AccessibilityButton, host.AccessibilityLink, host.AccessibilityHeading, host.AccessibilityParagraph, host.AccessibilityListItem, host.AccessibilityTextbox, host.AccessibilityCheckbox, host.AccessibilityRadio, host.AccessibilityCombobox:
		return true
	default:
		return false
	}
}
