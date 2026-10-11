package page

import (
	"testing"

	"github.com/chinmay-sawant/ownframe/internal/host"
)

func TestAccessibilityTreeBuildsNamedSemanticOrder(t *testing.T) {
	p := accessibilityTestPage(t, `<main id="main"><h1 id="welcome">Welcome</h1><label for="email">Email address</label><input id="email" value="a@b.test"><button id="save">Save</button><p hidden>Secret</p></main>`)
	tree, err := p.AccessibilityTree()
	if err != nil {
		t.Fatal(err)
	}
	main := accessibilityTestNode(t, tree, "id:main")
	if main.Role != host.AccessibilityMain || main.Parent != tree.Root {
		t.Fatalf("main: %+v", main)
	}
	if got := accessibilityTestNode(t, tree, "id:welcome").Name; got != "Welcome" {
		t.Fatalf("heading name %q", got)
	}
	field := accessibilityTestNode(t, tree, "id:email")
	if field.Name != "Email address" || field.Value != "a@b.test" || !field.HasBounds {
		t.Fatalf("field: %+v", field)
	}
	if accessibilityHasNode(tree, "Secret") {
		t.Fatal("hidden text was exposed")
	}
	if got := accessibilityTestNode(t, tree, "id:save").Name; got != "Save" {
		t.Fatalf("button name %q", got)
	}
}

func TestAccessibilityTreeRedactsAndDetaches(t *testing.T) {
	p := accessibilityTestPage(t, `<input id="password" type="password" value="secret"><button id="dup">A</button><button id="dup">B</button>`)
	tree, err := p.AccessibilityTree()
	if err != nil {
		t.Fatal(err)
	}
	password := accessibilityTestNode(t, tree, "id:password")
	if !password.Password || password.Value != "" {
		t.Fatalf("password: %+v", password)
	}
	for _, node := range tree.Nodes {
		if node.ElementID == "dup" && (node.HasBounds || len(node.Actions) != 0) {
			t.Fatalf("duplicate id is actionable: %+v", node)
		}
	}
	tree.Nodes[0].Children[0] = "changed"
	again, err := p.AccessibilityTree()
	if err != nil {
		t.Fatal(err)
	}
	if again.Nodes[0].Children[0] == "changed" {
		t.Fatal("snapshot aliases page state")
	}
}
