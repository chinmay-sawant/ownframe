package page

import (
	"context"
	"testing"

	"github.com/chinmay-sawant/ownframe/internal/host"
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
