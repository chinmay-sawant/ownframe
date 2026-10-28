package main

import (
	"os"
	"path/filepath"
	"testing"
)

func createLazyTestDir(t *testing.T) string {
	t.Helper()
	dir := t.TempDir()

	files := map[string]string{
		"main.go":                   "package main\n\nfunc main() {}\n",
		"README.md":                 "# Readme\n",
		"src/app.go":                "package src\n\nfunc App() {}\n",
		"src/view.go":               "package src\n\nfunc View() {}\n",
		"src/components/button.go":  "package components\n\nfunc Button() {}\n",
		"testdata/fixture.json":     `{"test": true}`,
		"internal/core/logic.go":    "package core\n\nfunc Logic() {}\n",
	}

	for path, content := range files {
		abs := filepath.Join(dir, filepath.FromSlash(path))
		if err := os.MkdirAll(filepath.Dir(abs), 0o700); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(abs, []byte(content), 0o600); err != nil {
			t.Fatal(err)
		}
	}

	return dir
}

func TestWorkspaceLazyInitialLoad(t *testing.T) {
	t.Parallel()

	dir := createLazyTestDir(t)

	ws, err := loadDiskWorkspaceLazy(dir)
	if err != nil {
		t.Fatalf("loadDiskWorkspaceLazy: %v", err)
	}

	if !ws.Lazy {
		t.Fatal("expected ws.Lazy to be true")
	}

	if ws.Truncated {
		t.Fatal("expected ws.Truncated to be false for lazy workspace")
	}

	if budget := ws.ScanBudget(); budget != "" {
		t.Fatalf("expected empty ScanBudget, got %q", budget)
	}

	// Root has 2 files: main.go and README.md.
	if len(ws.Files) != 2 {
		t.Fatalf("expected 2 root files, got %d", len(ws.Files))
	}

	if ws.ByPath["main.go"] == nil || ws.ByPath["README.md"] == nil {
		t.Fatalf("expected main.go and README.md in ByPath: %+v", ws.ByPath)
	}

	// Subdirectory files should NOT be loaded initially.
	if ws.ByPath["src/app.go"] != nil {
		t.Fatal("expected src/app.go to NOT be loaded yet")
	}

	// Check tree nodes in root: internal, src, testdata (dirs), main.go, README.md (files).
	dirs := map[string]*fileNode{}
	for _, node := range ws.Tree {
		if node.Dir {
			dirs[node.Path] = node
		}
	}

	for _, wantDir := range []string{"internal", "src", "testdata"} {
		node, ok := dirs[wantDir]
		if !ok {
			t.Fatalf("expected dir %q in root tree", wantDir)
		}
		if node.Loaded {
			t.Fatalf("expected dir %q to have Loaded=false initially", wantDir)
		}
		if len(node.Children) != 0 {
			t.Fatalf("expected dir %q to have empty children initially", wantDir)
		}
	}
}

func TestWorkspaceLazyLoadDir(t *testing.T) {
	t.Parallel()

	dir := createLazyTestDir(t)

	ws, err := loadDiskWorkspaceLazy(dir)
	if err != nil {
		t.Fatalf("loadDiskWorkspaceLazy: %v", err)
	}

	// Expand "src".
	if err := ws.LoadDir("src"); err != nil {
		t.Fatalf("ws.LoadDir(src): %v", err)
	}

	srcNode := findTreeNode(ws.Tree, "src")
	if srcNode == nil {
		t.Fatal("findTreeNode(src) returned nil")
	}

	if !srcNode.Loaded {
		t.Fatal("expected srcNode.Loaded to be true after LoadDir")
	}

	// In src: components (dir), app.go (file), view.go (file).
	if len(srcNode.Children) != 3 {
		t.Fatalf("expected 3 children in src, got %d", len(srcNode.Children))
	}

	// ws.Files now has 4 files (2 root + 2 in src).
	if len(ws.Files) != 4 {
		t.Fatalf("expected 4 files in ws.Files, got %d", len(ws.Files))
	}

	if ws.ByPath["src/app.go"] == nil || ws.ByPath["src/view.go"] == nil {
		t.Fatalf("expected src/app.go and src/view.go in ByPath")
	}

	// components child should still be unloaded.
	componentsNode := findTreeNode(ws.Tree, "src/components")
	if componentsNode == nil {
		t.Fatal("findTreeNode(src/components) returned nil")
	}
	if componentsNode.Loaded {
		t.Fatal("expected src/components.Loaded to be false before expanding it")
	}

	// Expand "src/components".
	if err := ws.LoadDir("src/components"); err != nil {
		t.Fatalf("ws.LoadDir(src/components): %v", err)
	}

	if !componentsNode.Loaded {
		t.Fatal("expected componentsNode.Loaded to be true")
	}

	if len(ws.Files) != 5 {
		t.Fatalf("expected 5 files after loading components, got %d", len(ws.Files))
	}

	if ws.ByPath["src/components/button.go"] == nil {
		t.Fatal("expected src/components/button.go in ByPath")
	}
}

func TestTreeLazyAppToggleDir(t *testing.T) {
	dir := createLazyTestDir(t)

	app, err := New(dir)
	if err != nil {
		t.Fatalf("New: %v", err)
	}

	if !app.ws.Lazy {
		t.Fatal("expected app.ws.Lazy to be true")
	}

	// Initial file counter and scan budget.
	v := app.View()
	if v.ScanBudget != "" {
		t.Fatalf("expected empty ScanBudget in view, got %q", v.ScanBudget)
	}
	if v.TreeCount != 2 {
		t.Fatalf("expected TreeCount=2 initially, got %d", v.TreeCount)
	}

	// Toggle "src" open.
	app.onToggleDir("src")

	if !app.folders["src"] {
		t.Fatal("expected app.folders[src] to be true")
	}

	srcNode := findTreeNode(app.ws.Tree, "src")
	if srcNode == nil || !srcNode.Loaded {
		t.Fatalf("expected src node to be loaded: %+v", srcNode)
	}

	// TreeCount should now be 4.
	v = app.View()
	if v.TreeCount != 4 {
		t.Fatalf("expected TreeCount=4 after expanding src, got %d", v.TreeCount)
	}

	// Check visible rows include src children.
	foundApp := false
	foundView := false
	foundComponents := false
	for _, row := range v.Tree {
		switch row.Path {
		case "src/app.go":
			foundApp = true
		case "src/view.go":
			foundView = true
		case "src/components":
			foundComponents = true
			if !row.Dir {
				t.Error("expected src/components to be a Dir row")
			}
		}
	}
	if !foundApp || !foundView || !foundComponents {
		t.Fatalf("missing rows after expanding src: app=%v view=%v components=%v", foundApp, foundView, foundComponents)
	}

	// Toggle "src" closed.
	app.onToggleDir("src")
	if app.folders["src"] {
		t.Fatal("expected app.folders[src] to be false after toggling again")
	}

	// Node remains loaded.
	if !srcNode.Loaded {
		t.Fatal("expected srcNode to remain Loaded after collapse")
	}

	// Rows should no longer show children of src.
	v = app.View()
	for _, row := range v.Tree {
		if row.Path == "src/app.go" || row.Path == "src/view.go" {
			t.Fatalf("unexpected child row %q visible after collapsing src", row.Path)
		}
	}
}

func TestTreeLazyFileNodeLoadChildrenDirect(t *testing.T) {
	t.Parallel()

	dir := createLazyTestDir(t)

	node := &fileNode{
		Path: "src",
		Name: "src",
		Dir:  true,
	}

	if err := node.LoadChildren(dir); err != nil {
		t.Fatalf("node.LoadChildren: %v", err)
	}

	if !node.Loaded {
		t.Fatal("expected node.Loaded to be true")
	}

	if len(node.Children) != 3 {
		t.Fatalf("expected 3 children, got %d", len(node.Children))
	}

	// First child should be the directory (components).
	if !node.Children[0].Dir || node.Children[0].Name != "components" {
		t.Fatalf("expected first child to be directory components, got %+v", node.Children[0])
	}
}

func TestTreeLazyEnsureFileLoadedOnOpen(t *testing.T) {
	dir := createLazyTestDir(t)

	app, err := New(dir)
	if err != nil {
		t.Fatalf("New: %v", err)
	}

	// File inside unexpanded directory: internal/core/logic.go
	if app.ws.ByPath["internal/core/logic.go"] != nil {
		t.Fatal("expected internal/core/logic.go to NOT be in ByPath initially")
	}

	app.openFile("internal/core/logic.go")

	if app.ws.ByPath["internal/core/logic.go"] == nil {
		t.Fatal("expected internal/core/logic.go to be loaded into ByPath by openFile")
	}

	found := false
	for _, tab := range app.tabs {
		if tab == "internal/core/logic.go" {
			found = true
			break
		}
	}
	if !found {
		t.Fatalf("expected open tab internal/core/logic.go, got tabs=%v", app.tabs)
	}

	doc := app.docs["internal/core/logic.go"]
	if doc == nil || doc.Err != "" {
		t.Fatalf("expected document to open without error, got %+v", doc)
	}
}
