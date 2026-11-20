package page

import (
	"context"
	"errors"
	"github.com/chinmay-sawant/ownframe/internal/host"
	"testing"
)

func TestAccessibilitySetValueUsesEditHooks(t *testing.T) {
	p := accessibilityTestPage(t, `<input id="email" value="old"><input id="check" type="checkbox">`)
	before, changed := 0, 0
	p.Handle(Handlers{BeforeEdit: func(context.Context, Box) error { before++; return nil }, Change: func(context.Context, Box) error { changed++; return nil }})
	tree, _ := p.AccessibilityTree()
	err := p.AccessibilityPerform(context.Background(), host.AccessibilityRequest{Document: tree.Document, Node: "id:email", Kind: host.AccessibilitySetValue, Value: "new@example.test"})
	if err != nil {
		t.Fatal(err)
	}
	if p.FormValue("email") != "new@example.test" || before != 1 || changed != 1 {
		t.Fatalf("value %q hooks %d/%d", p.FormValue("email"), before, changed)
	}
	tree, _ = p.AccessibilityTree()
	err = p.AccessibilityPerform(context.Background(), host.AccessibilityRequest{Document: tree.Document, Node: "id:check", Kind: host.AccessibilityActivate})
	if err != nil || !p.FormChecked("check") {
		t.Fatalf("activate: checked %v err %v", p.FormChecked("check"), err)
	}
	tree, _ = p.AccessibilityTree()
	check := accessibilityTestNode(t, tree, "id:check")
	if !check.Checked || !check.HasChecked {
		t.Fatalf("checkbox state: %+v", check)
	}
}

func TestAccessibilityAnonymousNodesGetUnambiguousBounds(t *testing.T) {
	p := accessibilityTestPage(t, `<main><p>Anonymous paragraph</p><button>Save</button><button>Save</button></main>`)
	tree, err := p.AccessibilityTree()
	if err != nil {
		t.Fatal(err)
	}
	var buttons int
	for _, node := range tree.Nodes {
		if node.Role != host.AccessibilityButton && node.Role != host.AccessibilityParagraph {
			continue
		}
		if !node.HasBounds {
			t.Fatalf("anonymous node lacks bounds: %+v", node)
		}
		if node.Role == host.AccessibilityButton {
			buttons++
		}
	}
	if buttons != 2 {
		t.Fatalf("anonymous buttons = %d, want 2", buttons)
	}
}

func TestAccessibilityAnonymousBoundsOmitAmbiguousCSSHiddenMatch(t *testing.T) {
	p := accessibilityTestPage(t, `<button style="display:none">Save</button><button>Save</button>`)
	tree, err := p.AccessibilityTree()
	if err != nil {
		t.Fatal(err)
	}
	for _, node := range tree.Nodes {
		if node.Role == host.AccessibilityButton && node.HasBounds {
			t.Fatalf("ambiguous anonymous node got bounds: %+v", node)
		}
	}
}

func TestAccessibilityRejectsStaleDocument(t *testing.T) {
	p := accessibilityTestPage(t, `<input id="email" value="old">`)
	tree, _ := p.AccessibilityTree()
	if err := p.Load(context.Background(), `<input id="email" value="new">`); err != nil {
		t.Fatal(err)
	}
	err := p.AccessibilityPerform(context.Background(), host.AccessibilityRequest{Document: tree.Document, Node: "id:email", Kind: host.AccessibilitySetValue, Value: "bad"})
	if !errors.Is(err, host.ErrAccessibilityStale) {
		t.Fatalf("stale error %v", err)
	}
}

func TestAccessibilityDisabledControlHasNoActions(t *testing.T) {
	p := accessibilityTestPage(t, `<button id="off" disabled>Off</button>`)
	tree, _ := p.AccessibilityTree()
	node := accessibilityTestNode(t, tree, "id:off")
	if !node.Disabled || len(node.Actions) != 0 {
		t.Fatalf("disabled node: %+v", node)
	}
	err := p.AccessibilityPerform(context.Background(), host.AccessibilityRequest{Document: tree.Document, Node: node.ID, Kind: host.AccessibilityFocus})
	if !errors.Is(err, host.ErrAccessibilityDisabled) {
		t.Fatalf("disabled error %v", err)
	}
}

func TestAccessibilityAriaDisabledCanTakeFocus(t *testing.T) {
	p := accessibilityTestPage(t, `<button id="off" aria-disabled="true">Off</button>`)
	tree, _ := p.AccessibilityTree()
	node := accessibilityTestNode(t, tree, "id:off")
	if !node.Disabled || !node.Focusable || len(node.Actions) != 1 {
		t.Fatalf("ARIA-disabled node: %+v", node)
	}
	err := p.AccessibilityPerform(context.Background(), host.AccessibilityRequest{Document: tree.Document, Node: node.ID, Kind: host.AccessibilityFocus})
	if err != nil || p.FocusID() != "off" {
		t.Fatalf("focus: id %q err %v", p.FocusID(), err)
	}
}

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
