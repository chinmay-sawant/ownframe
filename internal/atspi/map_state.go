package atspi

import "github.com/chinmay-sawant/ownframe/internal/host"

func mapState(n host.AccessibilityNode, focused host.AccessibilityNodeID) State {
	editable := false
	for _, kind := range n.Actions {
		if kind == host.AccessibilitySetValue {
			editable = true
			break
		}
	}
	return State{
		Enabled:   !n.Disabled,
		Sensitive: !n.Disabled,
		Focusable: n.Focusable,
		Focused:   n.Focused || n.ID == focused,
		Checkable: n.HasChecked,
		Checked:   n.Checked,
		Editable:  editable,
	}
}
