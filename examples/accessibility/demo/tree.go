package demo

import (
	"fmt"
	"strings"

	"github.com/chinmay-sawant/ownframe"
)

func treeText(tree ownframe.AccessibilityTree) string {
	var out strings.Builder
	fmt.Fprintf(&out, "Document=%d Revision=%d Root=%q Focused=%q Nodes=%d\n", tree.Document, tree.Revision, tree.Root, tree.Focused, len(tree.Nodes))
	for _, node := range tree.Nodes {
		value := node.Value
		if node.Password {
			value = "[redacted]"
		}
		actions := make([]string, len(node.Actions))
		for i, action := range node.Actions {
			actions[i] = string(action)
		}
		fmt.Fprintf(&out, "ID=%q Parent=%q Children=%q ElementID=%q Role=%s Name=%q Description=%q Value=%q Bounds=%+v HasBounds=%t Disabled=%t Focusable=%t Focused=%t Password=%t Checked=%t HasChecked=%t Actions=%v\n", node.ID, node.Parent, node.Children, node.ElementID, node.Role, node.Name, node.Description, value, node.Bounds, node.HasBounds, node.Disabled, node.Focusable, node.Focused, node.Password, node.Checked, node.HasChecked, actions)
	}
	return out.String()
}
