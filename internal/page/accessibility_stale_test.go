package page

import (
	"context"
	"errors"
	"github.com/chinmay-sawant/ownframe/internal/host"
	"testing"
)

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
