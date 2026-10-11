package demo

import (
	"context"
	"strings"
	"testing"

	"github.com/chinmay-sawant/ownframe"
)

func TestTreeAndSetValueAction(t *testing.T) {
	a, err := New()
	if err != nil {
		t.Fatal(err)
	}
	ctx := context.Background()
	if err := a.Page().Redraw(ctx); err != nil {
		t.Fatal(err)
	}
	if len(a.Page().PNG()) == 0 {
		t.Fatal("rendered PNG is empty")
	}
	tree, err := a.Page().AccessibilityTree()
	if err != nil {
		t.Fatal(err)
	}
	var name, secret ownframe.AccessibilityNode
	for _, node := range tree.Nodes {
		if node.ElementID == "display-name" {
			name = node
		}
		if node.ElementID == "secret" {
			secret = node
		}
	}
	if name.Name != "Display name" || name.Role != ownframe.AccessibilityTextbox {
		t.Fatalf("name node: %+v", name)
	}
	if !secret.Password || secret.Value != "" {
		t.Fatalf("password node: %+v", secret)
	}
	err = a.Page().AccessibilityPerform(ctx, ownframe.AccessibilityRequest{
		Document: tree.Document, Node: name.ID,
		Kind: ownframe.AccessibilitySetValue, Value: "Grace Hopper",
	})
	if err != nil {
		t.Fatal(err)
	}
	if got := a.Page().FormValue("display-name"); got != "Grace Hopper" {
		t.Fatalf("FormValue = %q", got)
	}
	if err := a.refreshTree(); err != nil {
		t.Fatal(err)
	}
	if strings.Contains(a.view.Tree, "private") || !strings.Contains(a.view.Tree, "[redacted]") {
		t.Fatalf("tree preview did not redact password: %q", a.view.Tree)
	}
}
