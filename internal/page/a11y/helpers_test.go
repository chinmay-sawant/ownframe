package a11y

import (
	"testing"

	"github.com/chinmay-sawant/blinkless/markup"
	"github.com/chinmay-sawant/ownframe/internal/host"
)

func TestHiddenAndPath(t *testing.T) {
	t.Parallel()

	if !Hidden(&markup.Node{Name: "script", Attrs: map[string]string{}}) {
		t.Fatal("script")
	}
	if !Hidden(&markup.Node{Name: "div", Attrs: map[string]string{"hidden": ""}}) {
		t.Fatal("hidden")
	}
	if !Hidden(&markup.Node{Name: "div", Attrs: map[string]string{"aria-hidden": "true"}}) {
		t.Fatal("aria-hidden")
	}
	if Hidden(&markup.Node{Name: "div", Attrs: map[string]string{}}) {
		t.Fatal("visible")
	}
	if ChildPath("", 0) != "0" || ChildPath("1", 2) != "1.2" {
		t.Fatal("path")
	}
}

func TestTextNamedAndActivate(t *testing.T) {
	t.Parallel()
	if !TextNamedByParent(host.AccessibilityButton) || TextNamedByParent(host.AccessibilityMain) {
		t.Fatal("named")
	}
	if !CanActivate("button", "") || !CanActivate("input", "checkbox") || CanActivate("div", "text") {
		t.Fatal("activate")
	}
}

func TestNodeTextSkipsScript(t *testing.T) {
	t.Parallel()
	n := &markup.Node{
		Type: markup.TypeElement, Name: "p", Attrs: map[string]string{},
		Children: []*markup.Node{
			{Type: markup.TypeText, Text: " Hello "},
			{Type: markup.TypeElement, Name: "script", Children: []*markup.Node{{Type: markup.TypeText, Text: "skip"}}},
			{Type: markup.TypeText, Text: " world "},
		},
	}
	if got := NodeText(n); got != "Hello world" {
		t.Fatalf("text %q", got)
	}
}
