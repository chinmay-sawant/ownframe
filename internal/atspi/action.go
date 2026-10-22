package atspi

import "github.com/chinmay-sawant/ownframe/internal/host"

const (
	ActionGrabFocus = "grab-focus"
	ActionClick     = "click"
	ActionSetText   = "set-text"
)

// MapActions copies advertised page actions to AT-SPI action names.
func MapActions(kinds []host.AccessibilityActionKind) []string {
	out := make([]string, 0, len(kinds))
	for _, kind := range kinds {
		if name, ok := actionName(kind); ok {
			out = append(out, name)
		}
	}
	return out
}

func actionName(kind host.AccessibilityActionKind) (string, bool) {
	switch kind {
	case host.AccessibilityFocus:
		return ActionGrabFocus, true
	case host.AccessibilityActivate:
		return ActionClick, true
	case host.AccessibilitySetValue:
		return ActionSetText, true
	default:
		return "", false
	}
}

// ActionKind maps an AT-SPI action name back to a page action.
func ActionKind(name string) (host.AccessibilityActionKind, bool) {
	switch name {
	case ActionGrabFocus:
		return host.AccessibilityFocus, true
	case ActionClick:
		return host.AccessibilityActivate, true
	case ActionSetText:
		return host.AccessibilitySetValue, true
	default:
		return 0, false
	}
}
