package page

import "github.com/chinmay-sawant/ownframe/internal/host"

func accessibilityNode(tree host.AccessibilityTree, id host.AccessibilityNodeID) (host.AccessibilityNode, bool) {
	for _, node := range tree.Nodes {
		if node.ID == id {
			return node, true
		}
	}
	return host.AccessibilityNode{}, false
}

func hasAccessibilityAction(node host.AccessibilityNode, want host.AccessibilityActionKind) bool {
	for _, action := range node.Actions {
		if action == want {
			return true
		}
	}
	return false
}
