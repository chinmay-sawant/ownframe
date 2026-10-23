package a11y

import (
	"testing"

	"github.com/chinmay-sawant/blinkless/markup"
	"github.com/chinmay-sawant/ownframe/internal/host"
)

func TestRoleTables(t *testing.T) {
	t.Parallel()

	el := func(name string, attrs map[string]string) *markup.Node {
		if attrs == nil {
			attrs = map[string]string{}
		}
		return &markup.Node{Type: markup.TypeElement, Name: name, Attrs: attrs}
	}

	cases := []struct {
		n    *markup.Node
		role host.AccessibilityRole
		ok   bool
	}{
		{el("main", nil), host.AccessibilityMain, true},
		{el("nav", nil), host.AccessibilityNavigation, true},
		{el("section", map[string]string{"aria-label": "box"}), host.AccessibilityGroup, true},
		{el("section", nil), "", false},
		{el("p", nil), host.AccessibilityParagraph, true},
		{el("a", map[string]string{"href": "/"}), host.AccessibilityLink, true},
		{el("a", nil), "", false},
		{el("img", map[string]string{"alt": "cat"}), host.AccessibilityImage, true},
		{el("img", nil), "", false},
		{el("button", nil), host.AccessibilityButton, true},
		{el("div", map[string]string{"role": "navigation"}), host.AccessibilityNavigation, true},
		{el("div", map[string]string{"role": "unknown"}), "", false},
	}
	for _, tc := range cases {
		role, ok := Role(tc.n)
		if role != tc.role || ok != tc.ok {
			t.Fatalf("%s attrs %v: %q %v", tc.n.Name, tc.n.Attrs, role, ok)
		}
	}
}

func TestInputRole(t *testing.T) {
	t.Parallel()
	role, ok := InputRole("checkbox")
	if role != host.AccessibilityCheckbox || !ok {
		t.Fatalf("checkbox %q %v", role, ok)
	}
	if _, ok := InputRole("hidden"); ok {
		t.Fatal("hidden")
	}
	role, ok = InputRole("")
	if role != host.AccessibilityTextbox || !ok {
		t.Fatalf("text %q %v", role, ok)
	}
}

func TestSupportedRoleEmpty(t *testing.T) {
	t.Parallel()
	if _, ok := SupportedRole(""); ok {
		t.Fatal("empty")
	}
	if _, ok := SupportedRole("  "); ok {
		t.Fatal("space")
	}
}
