package demo

import (
	"context"
	"testing"

	"github.com/chinmay-sawant/ownframe"
)

func TestExampleRunsEveryAccessibilityAction(t *testing.T) {
	a := renderedApp(t)
	ctx := context.Background()
	tree, err := a.accessible().AccessibilityTree()
	if err != nil {
		t.Fatal(err)
	}
	name := nodeByID(t, tree, "id:display-name")
	if err := a.accessible().AccessibilityPerform(ctx, ownframe.AccessibilityRequest{
		Document: tree.Document, Node: name.ID, Kind: ownframe.AccessibilityFocus,
	}); err != nil {
		t.Fatal(err)
	}
	if got := a.Page().Form().FocusID(); got != "display-name" {
		t.Fatalf("focused id = %q", got)
	}
	focused, err := a.accessible().AccessibilityTree()
	if err != nil || focused.Focused != name.ID || !nodeByID(t, focused, "id:display-name").Focused {
		t.Fatalf("focused tree node: tree=%+v error=%v", focused, err)
	}
	if err := a.performElement(ctx, "display-name", ownframe.AccessibilitySetValue, "Grace Hopper"); err != nil {
		t.Fatal(err)
	}
	if got := a.Page().Form().FormValue("display-name"); got != "Grace Hopper" {
		t.Fatalf("name value = %q", got)
	}
	if err := a.performElement(ctx, "updates", ownframe.AccessibilityActivate, ""); err != nil {
		t.Fatal(err)
	}
	if a.Page().Form().FormChecked("updates") {
		t.Fatal("checkbox was not toggled")
	}
	if err := a.performElement(ctx, "contact-phone", ownframe.AccessibilityActivate, ""); err != nil {
		t.Fatal(err)
	}
	if !a.Page().Form().FormChecked("contact-phone") || a.Page().Form().FormChecked("contact-email") {
		t.Fatal("radio not selected")
	}
	oldPlan := a.Page().Form().FormValue("plan")
	if err := a.performElement(ctx, "plan", ownframe.AccessibilityActivate, ""); err != nil {
		t.Fatal(err)
	}
	if got := a.Page().Form().FormValue("plan"); got == oldPlan {
		t.Fatalf("select value did not change: %q", got)
	}
	if err := a.performElement(ctx, "activation-target", ownframe.AccessibilityActivate, ""); err != nil {
		t.Fatal(err)
	}
	if a.view.ActivationCount != 1 {
		t.Fatalf("activation count = %d", a.view.ActivationCount)
	}
}
