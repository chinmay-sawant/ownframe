package demo

import (
	"context"

	"github.com/chinmay-sawant/ownframe"
)

func (a *App) performElement(ctx context.Context, element string, kind ownframe.AccessibilityActionKind, value string) error {
	tree, err := a.accessible().AccessibilityTree()
	if err != nil {
		return err
	}
	for _, node := range tree.Nodes {
		if node.ElementID == element {
			return a.accessible().AccessibilityPerform(ctx, ownframe.AccessibilityRequest{
				Document: tree.Document, Node: node.ID, Kind: kind, Value: value,
			})
		}
	}
	return ownframe.ErrAccessibilityNode
}

func (a *App) runErrorProbe(ctx context.Context, probe string) error {
	tree, err := a.accessible().AccessibilityTree()
	if err != nil {
		return err
	}
	r := ownframe.AccessibilityRequest{Document: tree.Document}
	switch probe {
	case "probe-stale":
		r.Document++
		r.Node, r.Kind = "id:display-name", ownframe.AccessibilityFocus
	case "probe-node":
		r.Node, r.Kind = "id:missing", ownframe.AccessibilityFocus
	case "probe-action":
		r.Node, r.Kind = "id:display-name", ownframe.AccessibilityActivate
	case "probe-disabled":
		r.Node, r.Kind = "id:disabled-field", ownframe.AccessibilityFocus
	case "probe-range":
		r.Node, r.Kind, r.Value = "id:display-name", ownframe.AccessibilitySetValue, string([]byte{0xff})
	}
	return a.accessible().AccessibilityPerform(ctx, r)
}
