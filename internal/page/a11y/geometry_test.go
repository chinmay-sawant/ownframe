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
