package demo

import (
	"testing"

	"github.com/chinmay-sawant/ownframe"
)

func TestTreeExposesNodeMetadataAndActions(t *testing.T) {
	tree, err := renderedApp(t).accessible().AccessibilityTree()
	if err != nil {
		t.Fatal(err)
	}
	if tree.Document == 0 || tree.Revision == 0 || tree.Root != "document" || len(tree.Nodes) == 0 {
		t.Fatalf("tree metadata: %+v", tree)
	}
	main := nodeByID(t, tree, "id:content-main")
	if main.Role != ownframe.AccessibilityMain || main.Parent != tree.Root || len(main.Children) == 0 {
		t.Fatalf("main tree links: %+v", main)
	}
	name := nodeByID(t, tree, "id:display-name")
	if name.ElementID != "display-name" || name.Name != "Display name" || name.Description == "" || name.Value != "Ada" || !name.HasBounds || name.Bounds.W <= 0 || name.Bounds.H <= 0 || !name.Focusable {
		t.Fatalf("textbox metadata: %+v", name)
	}
	if !hasAction(name, ownframe.AccessibilityFocus) || !hasAction(name, ownframe.AccessibilitySetValue) {
		t.Fatalf("textbox actions: %v", name.Actions)
	}
	check := nodeByID(t, tree, "id:updates")
	if !check.Checked || !check.HasChecked || !hasAction(check, ownframe.AccessibilityActivate) {
		t.Fatalf("checkbox metadata: %+v", check)
	}
	if !nodeByID(t, tree, "id:secret").Password {
		t.Fatal("password metadata is missing")
	}
	disabled := nodeByID(t, tree, "id:disabled-field")
	if !disabled.Disabled || disabled.Focusable || len(disabled.Actions) != 0 {
		t.Fatal("disabled state is missing")
	}
	if nodeByID(t, tree, "id:docs-link").Name != "Action controls" || nodeByID(t, tree, "id:avatar").Name != "Blue profile avatar" {
		t.Fatal("link or image name is missing")
	}
	if nodeByID(t, tree, "id:semantics").Name != "Semantic fixtures" {
		t.Fatal("aria-labelledby group name is missing")
	}
}

func hasAction(node ownframe.AccessibilityNode, want ownframe.AccessibilityActionKind) bool {
	for _, action := range node.Actions {
		if action == want {
			return true
		}
	}
	return false
}
