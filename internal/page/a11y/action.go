package a11y

import "github.com/chinmay-sawant/ownframe/internal/host"

// Lookup finds the node with id in tree.
func Lookup(tree host.AccessibilityTree, id host.AccessibilityNodeID) (host.AccessibilityNode, bool) {
	for _, node := range tree.Nodes {
		if node.ID == id {
			return node, true
		}
	}
	return host.AccessibilityNode{}, false
}

// HasAction reports that node advertises want.
func HasAction(node host.AccessibilityNode, want host.AccessibilityActionKind) bool {
	for _, action := range node.Actions {
		if action == want {
			return true
		}
	}
	return false
}
