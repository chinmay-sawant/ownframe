package atspi

import (
	"github.com/chinmay-sawant/ownframe/internal/host"
	"slices"
)

func mapNode(n host.AccessibilityNode, paths map[host.AccessibilityNodeID]string, focused host.AccessibilityNodeID) Object {
	obj := Object{
		Path:        paths[n.ID],
		Parent:      paths[n.Parent],
		ID:          n.ID,
		Role:        MapRole(n.Role, n.Password),
		Name:        n.Name,
		Description: n.Description,
		Value:       n.Value,
		State:       mapState(n, focused),
		Actions:     MapActions(n.Actions),
		Bounds:      n.Bounds,
		HasBounds:   n.HasBounds,
		Password:    n.Password,
	}
	if n.Password {
		obj.Value = ""
		obj.State.Editable = false
		obj.Actions = slices.DeleteFunc(obj.Actions, func(action string) bool { return action == ActionSetText })
	}
	obj.Children = childPaths(n.Children, paths)
	return obj
}

func childPaths(ids []host.AccessibilityNodeID, paths map[host.AccessibilityNodeID]string) []string {
	out := make([]string, 0, len(ids))
	for _, id := range ids {
		if p := paths[id]; p != "" {
			out = append(out, p)
		}
	}
	return out
}
