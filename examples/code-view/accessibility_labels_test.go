package main

import (
	"context"
	"strings"
	"testing"

	"github.com/chinmay-sawant/ownframe"
)

const (
	paletteFocusSteps = 12
	branchFocusSteps  = 8
)

// TestAccessibilityLabels checks that every focusable control in the shell
// carries an accessible name, including the icon-only buttons and the find
// bar controls.
func TestAccessibilityLabels(t *testing.T) {
	app := newTestApp(t)

	tree := accessibilityTree(t, app)

	assertFocusableNames(t, tree)

	want := map[string]string{
		"act-explorer": "Explorer",
		"act-changes":  "Source control",
		"act-search":   "Search workspace",
		"act-details":  "Details closed",
		"act-collapse": "Collapse all folders",
		"act-devtools": "Developer tools on",
		"tab-details":  "Details closed",
		"wrap-toggle":  "Soft wrap on",
	}

	for id, name := range want {
		assertAccessibilityName(t, tree, id, name)
	}

	if err := app.openFind(context.Background()); err != nil {
		t.Fatalf("open find: %v", err)
	}

	tree = accessibilityTree(t, app)

	assertFocusableNames(t, tree)

	find := map[string]string{
		"find-case":  "Match case",
		"find-prev":  "Previous match",
		"find-next":  "Next match",
		"find-close": "Close find",
	}

	for id, name := range find {
		assertAccessibilityName(t, tree, id, name)
	}

	assertAccessibilityName(t, tree, "find", "Find in file")
}

// TestAccessibilityPaletteDialog checks the quick-open overlay exposes a
// named group and that both overlays declare themselves as dialogs.
func TestAccessibilityPaletteDialog(t *testing.T) {
	app := newTestApp(t)

	if err := app.openPalette(context.Background()); err != nil {
		t.Fatalf("open palette: %v", err)
	}

	tree := accessibilityTree(t, app)

	namedGroup := false

	for _, node := range tree.Nodes {
		if node.Role == ownframe.AccessibilityGroup && node.Name == "Quick open results" {
			namedGroup = true
		}
	}

	if !namedGroup {
		t.Fatal("palette is open but the tree has no named quick-open group")
	}

	assertAccessibilityName(t, tree, "palette", "Search files")

	markup := buildHTML()
	if got := strings.Count(markup, `role="dialog"`); got < 2 {
		t.Errorf(`role="dialog" count = %d, want at least 2`, got)
	}

	if got := strings.Count(markup, `aria-modal="true"`); got < 2 {
		t.Errorf(`aria-modal="true" count = %d, want at least 2`, got)
	}

	if !strings.Contains(markup, `aria-live="polite"`) {
		t.Error(`missing aria-live="polite"`)
	}
}

// TestAccessibilityChangesMode exercises the branch dialog and diff controls,
// which are only rendered outside the explorer view.
func TestAccessibilityChangesMode(t *testing.T) {
	app := newTestApp(t)
	ctx := context.Background()

	app.setMode(ctx, "changes")
	app.branchOpen = true
	app.refresh()

	tree := accessibilityTree(t, app)

	assertFocusableNames(t, tree)

	want := map[string]string{
		"tab-details":   "Details closed",
		"hunk-prev":     "Previous hunk",
		"hunk-next":     "Next hunk",
		"seg-split":     "Split",
		"branch-filter": "Filter branches",
		"branch-head":   "Follow HEAD",
		"branch-close":  "Close",
	}

	for id, name := range want {
		assertAccessibilityName(t, tree, id, name)
	}

	app.view.Modal = true
	app.page.SetData(&app.view)

	if err := app.page.Redraw(ctx); err != nil {
		t.Fatalf("redraw: %v", err)
	}

	if err := app.page.Form().Focus(ctx, "branch-filter"); err != nil {
		t.Fatalf("focus branch filter: %v", err)
	}

	for i := 0; i < branchFocusSteps; i++ {
		if err := app.page.Form().FocusNext(ctx); err != nil {
			t.Fatalf("focus next: %v", err)
		}

		id := app.page.Form().FocusID()
		if !strings.HasPrefix(id, "branch-") {
			t.Fatalf("focus left the branch dialog: %q", id)
		}
	}
}

// TestAccessibilityFocusTrap checks that the Modal flag removes the background
// controls from the tab order, so focus wraps inside the palette overlay.
func TestAccessibilityFocusTrap(t *testing.T) {
	app := newTestApp(t)
	ctx := context.Background()

	if err := app.openPalette(ctx); err != nil {
		t.Fatalf("open palette: %v", err)
	}

	app.view.Modal = true
	app.page.SetData(&app.view)

	if err := app.page.Redraw(ctx); err != nil {
		t.Fatalf("redraw: %v", err)
	}

	if err := app.page.Form().Focus(ctx, "palette"); err != nil {
		t.Fatalf("focus palette: %v", err)
	}

	for i := 0; i < paletteFocusSteps; i++ {
		if err := app.page.Form().FocusNext(ctx); err != nil {
			t.Fatalf("focus next: %v", err)
		}

		id := app.page.Form().FocusID()
		if id != "palette" && id != "palette-close" && !strings.HasPrefix(id, "palette-") {
			t.Fatalf("focus left the palette overlay: %q", id)
		}
	}
}

func assertFocusableNames(t *testing.T, tree ownframe.AccessibilityTree) {
	t.Helper()

	for _, node := range tree.Nodes {
		if node.Focusable && strings.TrimSpace(node.Name) == "" {
			t.Errorf("focusable %s %q has no name", node.Role, node.ElementID)
		}
	}
}

func assertAccessibilityName(t *testing.T, tree ownframe.AccessibilityTree, id, name string) {
	t.Helper()

	node, ok := accessibilityNode(tree, id)
	if !ok {
		t.Errorf("%s is missing from the accessibility tree", id)

		return
	}

	if !node.Focusable {
		t.Errorf("%s is not focusable", id)
	}

	if node.Name != name {
		t.Errorf("%s name = %q, want %q", id, node.Name, name)
	}
}

func accessibilityTree(t *testing.T, app *App) ownframe.AccessibilityTree {
	t.Helper()

	if err := app.Page().Redraw(context.Background()); err != nil {
		t.Fatalf("redraw: %v", err)
	}

	tree, err := app.Page().Access().AccessibilityTree()
	if err != nil {
		t.Fatalf("accessibility tree: %v", err)
	}

	if len(tree.Nodes) == 0 {
		t.Fatal("accessibility tree is empty")
	}

	return tree
}

func accessibilityNode(tree ownframe.AccessibilityTree, id string) (ownframe.AccessibilityNode, bool) {
	for _, node := range tree.Nodes {
		if node.ElementID == id {
			return node, true
		}
	}

	return ownframe.AccessibilityNode{}, false
}
