package a11y

import (
	"testing"

	"github.com/chinmay-sawant/blinkless/markup"
	"github.com/chinmay-sawant/ownframe/internal/host"
)

func TestNamesFromLabelAndARIA(t *testing.T) {
	t.Parallel()

	label := &markup.Node{
		Type: markup.TypeElement, Name: "label", Attrs: map[string]string{"for": "email"},
		Children: []*markup.Node{{Type: markup.TypeText, Text: "Email"}},
	}
	field := &markup.Node{Type: markup.TypeElement, Name: "input", Attrs: map[string]string{"id": "email"}}
	hint := &markup.Node{
		Type: markup.TypeElement, Name: "span", Attrs: map[string]string{"id": "hint"},
		Children: []*markup.Node{{Type: markup.TypeText, Text: "Required"}},
	}
	root := &markup.Node{
		Type: markup.TypeElement, Name: "form", Attrs: map[string]string{},
		Children: []*markup.Node{label, field, hint},
	}
	names := NewNames()
	names.Index(root, nil)

	if got := names.Name(field, host.AccessibilityTextbox); got != "Email" {
		t.Fatalf("label %q", got)
	}
	field.Attrs["aria-labelledby"] = "hint"
	if got := names.Name(field, host.AccessibilityTextbox); got != "Required" {
		t.Fatalf("labelledby %q", got)
	}
	if got := names.Description(&markup.Node{Attrs: map[string]string{"aria-describedby": "hint"}}); got != "Required" {
		t.Fatalf("describedby %q", got)
	}
}

func TestNamesARIALabelAndTitle(t *testing.T) {
	t.Parallel()
	names := NewNames()
	btn := &markup.Node{Type: markup.TypeElement, Name: "button", Attrs: map[string]string{"aria-label": "Go"}}
	if got := names.Name(btn, host.AccessibilityButton); got != "Go" {
		t.Fatalf("aria-label %q", got)
	}
	img := &markup.Node{Type: markup.TypeElement, Name: "img", Attrs: map[string]string{"alt": "cat"}}
	if got := names.Name(img, host.AccessibilityImage); got != "cat" {
		t.Fatalf("alt %q", got)
	}
}
