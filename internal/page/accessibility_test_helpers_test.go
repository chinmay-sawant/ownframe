package page

import (
	"context"
	"testing"

	"github.com/chinmay-sawant/ownframe/internal/host"
)

func accessibilityTestPage(t *testing.T, source string) *Page {
	t.Helper()
	p, err := New(Config{HTML: source, Width: 400, Height: 240})
	if err != nil {
		t.Fatal(err)
	}
	if err := p.Redraw(context.Background()); err != nil {
		t.Fatal(err)
	}
	return p
}

func accessibilityTestNode(t *testing.T, tree host.AccessibilityTree, id host.AccessibilityNodeID) host.AccessibilityNode {
	t.Helper()
	for _, node := range tree.Nodes {
		if node.ID == id {
			return node
		}
	}
	t.Fatalf("node %q missing from %+v", id, tree.Nodes)
	return host.AccessibilityNode{}
}

func accessibilityHasNode(tree host.AccessibilityTree, name string) bool {
	for _, node := range tree.Nodes {
		if node.Name == name {
			return true
		}
	}
	return false
}
