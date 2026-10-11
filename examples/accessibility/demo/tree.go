package demo

import (
	"context"
	"fmt"
	"strings"

	"github.com/chinmay-sawant/ownframe"
)

func (a *App) setNameViaTree(ctx context.Context) error {
	tree, err := a.page.AccessibilityTree()
	if err != nil {
		return err
	}
	for _, node := range tree.Nodes {
		if node.ElementID == "display-name" {
			return a.page.AccessibilityPerform(ctx, ownframe.AccessibilityRequest{
				Document: tree.Document, Node: node.ID,
				Kind: ownframe.AccessibilitySetValue, Value: "Updated through tree action",
			})
		}
	}
	return ownframe.ErrAccessibilityNode
}

func treeText(tree ownframe.AccessibilityTree) string {
	var out strings.Builder
	for _, node := range tree.Nodes {
		if node.Name == "" && node.Value == "" && !node.Password {
			continue
		}
		value := strings.ReplaceAll(node.Value, "\n", " ")
		if node.Password {
			value = "[redacted]"
		}
		fmt.Fprintf(&out, "%s  %s", node.Role, node.Name)
		if node.Focused {
			out.WriteString("  [focused]")
		}
		if node.HasChecked {
			fmt.Fprintf(&out, "  checked=%t", node.Checked)
		}
		if value != "" {
			fmt.Fprintf(&out, "  value=%q", value)
		}
		out.WriteByte('\n')
	}
	return out.String()
}
