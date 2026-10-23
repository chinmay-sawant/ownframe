package demo

import (
	"context"
	"strings"
	"testing"
)

func TestPageActionButtonsReachTheAccessibilityAPI(t *testing.T) {
	a := renderedApp(t)
	click := func(id string) {
		t.Helper()
		tree, err := a.accessible().AccessibilityTree()
		if err != nil {
			t.Fatal(err)
		}
		node := nodeByID(t, tree, "id:"+id)
		if !node.HasBounds {
			t.Fatalf("button %q has no bounds", id)
		}
		err = a.Page().Click(
			context.Background(),
			node.Bounds.X+node.Bounds.W/2,
			node.Bounds.Y+node.Bounds.H/2,
		)
		if err != nil {
			t.Fatal(err)
		}
	}
	click("api-set-value")
	if got := a.Page().Form().FormValue("display-name"); got != "Updated through the tree action" {
		t.Fatalf("button SetValue result = %q", got)
	}
	click("probe-stale")
	if !strings.Contains(a.view.Status, "ErrAccessibilityStale") {
		t.Fatalf("error probe status = %q", a.view.Status)
	}
}
