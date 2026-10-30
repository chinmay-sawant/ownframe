package main

import (
	"context"
	"fmt"
	"testing"

	"github.com/chinmay-sawant/ownframe"
)

func manyStickyDir(t *testing.T) (*App, string) {
	t.Helper()

	dir := t.TempDir()
	writeFile(t, dir, "README.md", "# Fixture\n")

	for i := 0; i < 60; i++ {
		writeFile(t, dir, fmt.Sprintf("many/file%02d.go", i), "package many\n")
	}

	app, err := New(dir)
	if err != nil {
		t.Fatal(err)
	}

	app.Page().SetSize(1280, 720)

	if err := app.Page().Redraw(context.Background()); err != nil {
		t.Fatal(err)
	}

	clickID(t, app, "tree-many")

	return app, dir
}

func TestStickyAncestorVisibleAfterScroll(t *testing.T) {
	app, _ := manyStickyDir(t)

	app.scrollTree(5)

	view := app.View()
	if len(view.StickyTree) != 1 || view.StickyTree[0].Path != "many" {
		t.Fatalf("sticky = %+v", view.StickyTree)
	}

	if !view.StickyTree[0].Dir || !view.StickyTree[0].Open {
		t.Fatalf("sticky row is not an open dir: %+v", view.StickyTree[0])
	}

	for _, row := range view.Tree {
		if row.Path == "many" {
			t.Fatal("pinned ancestor must not repeat in the scroll window")
		}
	}

	if view.TreeVisible+len(view.StickyTree) != app.treeCapacity() {
		t.Fatalf("visible %d + sticky %d != capacity %d", view.TreeVisible, len(view.StickyTree), app.treeCapacity())
	}
}

func TestStickyClickCollapses(t *testing.T) {
	app, _ := manyStickyDir(t)
	ctx := context.Background()

	app.scrollTree(5)

	if len(app.View().StickyTree) == 0 {
		t.Fatal("expected a sticky row before clicking")
	}

	if err := app.onClick(ctx, ownframe.Box{Action: "dir:many"}); err != nil {
		t.Fatal(err)
	}

	if app.folders["many"] {
		t.Fatalf("many still open: %v", app.folders)
	}

	if len(app.View().StickyTree) != 0 {
		t.Fatalf("sticky after collapse = %+v", app.View().StickyTree)
	}
}

func TestStickyNestedStack(t *testing.T) {
	dir := t.TempDir()
	writeFile(t, dir, "README.md", "# Fixture\n")

	for i := 0; i < 40; i++ {
		writeFile(t, dir, fmt.Sprintf("a/b/c/file%02d.go", i), "package c\n")
	}

	app, err := New(dir)
	if err != nil {
		t.Fatal(err)
	}

	app.Page().SetSize(1280, 720)

	if err := app.Page().Redraw(context.Background()); err != nil {
		t.Fatal(err)
	}

	ctx := context.Background()
	for _, path := range []string{"a", "a/b", "a/b/c"} {
		if err := app.onClick(ctx, ownframe.Box{Action: "dir:" + path}); err != nil {
			t.Fatalf("open %s: %v", path, err)
		}
	}

	index := 0
	if !treeRowIndex(app.ws.Tree, app.folders, "a/b/c/file20.go", &index) {
		t.Fatal("no index for nested file")
	}

	app.treeOffset = index
	app.refresh()

	view := app.View()
	if len(view.StickyTree) != 3 {
		t.Fatalf("sticky stack = %+v", view.StickyTree)
	}

	want := []string{"a", "a/b", "a/b/c"}
	for i, path := range want {
		if view.StickyTree[i].Path != path {
			t.Fatalf("sticky[%d] = %q, want %q", i, view.StickyTree[i].Path, path)
		}
	}
}

func TestStickyCollapseClears(t *testing.T) {
	app, _ := manyStickyDir(t)

	app.scrollTree(5)

	if len(app.View().StickyTree) == 0 {
		t.Fatal("expected sticky before collapse")
	}

	app.collapseAll()

	if len(app.View().StickyTree) != 0 {
		t.Fatalf("sticky after collapseAll = %+v", app.View().StickyTree)
	}

	if app.treeOffset != 0 {
		t.Fatalf("offset after collapse = %d", app.treeOffset)
	}
}

func TestStickySkippedWhenFiltered(t *testing.T) {
	app, _ := manyStickyDir(t)

	app.scrollTree(5)

	if len(app.View().StickyTree) == 0 {
		t.Fatal("expected sticky before filtering")
	}

	app.filter = "file01"
	app.refresh()

	if len(app.View().StickyTree) != 0 {
		t.Fatalf("sticky with filter = %+v", app.View().StickyTree)
	}

	if !app.View().Filtered {
		t.Fatal("expected filtered view")
	}
}

func TestSpaceExpandsAndCollapsesDir(t *testing.T) {
	dir := createLazyTestDir(t)

	app, err := New(dir)
	if err != nil {
		t.Fatal(err)
	}

	if err := app.Page().Redraw(context.Background()); err != nil {
		t.Fatal(err)
	}

	ctx := context.Background()

	if err := app.Page().Form().Focus(ctx, "tree-src"); err != nil {
		t.Fatal(err)
	}

	if err := app.onKeyDown(ctx, "space"); err != nil {
		t.Fatal(err)
	}

	if !app.folders["src"] {
		t.Fatal("space did not expand src")
	}

	if node := findTreeNode(app.ws.Tree, "src"); node == nil || !node.Loaded {
		t.Fatal("space expand did not lazy-load src")
	}

	if got := app.Page().Form().FocusID(); got != "tree-src" {
		t.Fatalf("focus after expand = %q", got)
	}

	if err := app.onKeyDown(ctx, "space"); err != nil {
		t.Fatal(err)
	}

	if app.folders["src"] {
		t.Fatal("second space did not collapse src")
	}

	if got := app.Page().Form().FocusID(); got != "tree-src" {
		t.Fatalf("focus after collapse = %q", got)
	}
}

func TestSpaceOnFileOpens(t *testing.T) {
	app := newTestApp(t)
	ctx := context.Background()

	if err := app.Page().Form().Focus(ctx, "tree-backend/main.go"); err != nil {
		t.Fatal(err)
	}

	if err := app.onKeyDown(ctx, "space"); err != nil {
		t.Fatal(err)
	}

	if app.View().Active != "backend/main.go" {
		t.Fatalf("active = %q", app.View().Active)
	}

	if got := app.Page().Form().FocusID(); got != "tree-backend/main.go" {
		t.Fatalf("focus after open = %q", got)
	}
}

func TestSpaceInFilterStillTypes(t *testing.T) {
	app := newTestApp(t)
	ctx := context.Background()

	if err := app.Page().Form().Focus(ctx, "filter"); err != nil {
		t.Fatal(err)
	}

	before := len(app.folders)

	if err := app.onKeyDown(ctx, "space"); err != nil {
		t.Fatal(err)
	}

	if len(app.folders) != before {
		t.Fatalf("space in filter toggled folders: %v", app.folders)
	}

	if app.View().Active == "" && len(app.View().Tree) == 0 {
		t.Fatal("filter view broke after space")
	}
}
