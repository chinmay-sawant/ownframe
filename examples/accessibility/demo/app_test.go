package demo

import (
	"context"
	"strings"
	"testing"

	"github.com/chinmay-sawant/ownframe"
)

func renderedApp(t *testing.T) *App {
	t.Helper()
	a, err := New()
	if err != nil {
		t.Fatal(err)
	}
	if err := a.Page().RedrawIfChanged(context.Background()); err != nil {
		t.Fatal(err)
	}
	return a
}

func nodeByID(t *testing.T, tree ownframe.AccessibilityTree, id string) ownframe.AccessibilityNode {
	t.Helper()
	for _, node := range tree.Nodes {
		if node.ID == ownframe.AccessibilityNodeID(id) {
			return node
		}
	}
	t.Fatalf("node %q missing", id)
	return ownframe.AccessibilityNode{}
}

func TestTreePreviewRedactsPassword(t *testing.T) {
	a := renderedApp(t)
	tree, err := a.accessible().AccessibilityTree()
	if err != nil {
		t.Fatal(err)
	}
	secret := nodeByID(t, tree, "id:secret")
	if !secret.Password || secret.Value != "" {
		t.Fatalf("password node: %+v", secret)
	}
	preview := treeText(tree)
	if strings.Contains(preview, "private") || !strings.Contains(preview, "[redacted]") {
		t.Fatalf("password preview: %q", preview)
	}
}
