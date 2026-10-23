package a11y

import (
	"testing"

	"github.com/chinmay-sawant/blinkless/layout"
	"github.com/chinmay-sawant/blinkless/markup"
	"github.com/chinmay-sawant/ownframe/internal/host"
)

func TestUniqueBoxesAndLookup(t *testing.T) {
	t.Parallel()

	byID, counts := UniqueBoxes([]layout.Box{
		{ID: "a", Tag: "p"},
		{ID: "a", Tag: "p"},
		{ID: "b", Tag: "button"},
		{Tag: "span"},
	})
	if counts["a"] != 2 || counts["b"] != 1 || byID["b"].Tag != "button" {
		t.Fatalf("unique %+v %+v", byID, counts)
	}

	tree := host.AccessibilityTree{Nodes: []host.AccessibilityNode{
		{ID: "id:a", Actions: []host.AccessibilityActionKind{host.AccessibilityFocus}},
	}}
	node, ok := Lookup(tree, "id:a")
	if !ok || !HasAction(node, host.AccessibilityFocus) || HasAction(node, host.AccessibilityActivate) {
		t.Fatalf("lookup %+v %v", node, ok)
	}
	if _, ok := Lookup(tree, "missing"); ok {
		t.Fatal("missing")
	}
}

func TestNodeBoxesMatchUnambiguous(t *testing.T) {
	t.Parallel()
	p := &markup.Node{
		Type: markup.TypeElement, Name: "p", Attrs: map[string]string{},
		Children: []*markup.Node{{Type: markup.TypeText, Text: "Hi"}},
	}
	root := &markup.Node{Type: markup.TypeElement, Name: "main", Attrs: map[string]string{}, Children: []*markup.Node{p}}
	out := NodeBoxes(root, []layout.Box{{Tag: "p", Text: "Hi", X: 1}})
	if got, ok := out[p]; !ok || got.X != 1 {
		t.Fatalf("boxes %+v", out)
	}
	out = NodeBoxes(root, []layout.Box{{Tag: "p", Text: "Hi"}, {Tag: "p", Text: "Hi"}})
	if _, ok := out[p]; ok {
		t.Fatal("ambiguous")
	}
}

func TestIDPrefersUnique(t *testing.T) {
	t.Parallel()
	n := &markup.Node{Attrs: map[string]string{"id": "x"}}
	ids := map[string][]*markup.Node{"x": {n}}
	if ID(n, "0", ids) != "id:x" {
		t.Fatal("unique id")
	}
	ids["x"] = []*markup.Node{n, n}
	if ID(n, "0", ids) != "path:0" {
		t.Fatal("dup id")
	}
}

func TestAnonymousControlsPreserveDeclaredState(t *testing.T) {
	root, err := markup.Parse([]byte(`<main><input type="checkbox" checked><input type="radio"><input type="password" value="private"><textarea>draft</textarea></main>`))
	if err != nil {
		t.Fatal(err)
	}
	tree := build(root, View{})
	found := map[host.AccessibilityRole]int{}
	passwords, drafts := 0, 0
	for _, node := range tree.Nodes {
		found[node.Role]++
		if node.Role == host.AccessibilityCheckbox && (!node.HasChecked || !node.Checked) {
			t.Fatal("anonymous checkbox lost checked state")
		}
		if node.Role == host.AccessibilityRadio && (!node.HasChecked || node.Checked) {
			t.Fatal("anonymous radio lost unchecked state")
		}
		if node.Password {
			passwords++
			if node.Value != "" {
				t.Fatal("anonymous password leaked a value")
			}
		}
		if node.Value == "draft" {
			drafts++
		}
		if node.ElementID == "" && len(node.Actions) != 0 {
			t.Fatal("anonymous node advertised an unsupported action")
		}
	}
	if found[host.AccessibilityCheckbox] != 1 || found[host.AccessibilityRadio] != 1 || passwords != 1 || drafts != 1 {
		t.Fatal("control nodes missing")
	}
}
