package main

import (
	"context"
	"os/exec"
	"testing"
)

// TestPaletteInMemoryMatchesAppearFirst verifies that in-memory matches from
// open tabs, recent files, and existing workspace files appear before any disk
// fallback matches.
func TestPaletteInMemoryMatchesAppearFirst(t *testing.T) {
	dir := t.TempDir()
	writeFile(t, dir, "README.md", "# Test Project\n")
	writeFile(t, dir, "pkg/in_memory_service.go", "package pkg\n")
	writeFile(t, dir, "pkg/tab_service.go", "package pkg\n")

	app, err := New(dir)
	if err != nil {
		t.Fatalf("New: %v", err)
	}

	// Open tab_service.go so it becomes an open tab.
	app.openFile("pkg/tab_service.go")

	// Write a new file to disk that is not in the in-memory workspace.
	writeFile(t, dir, "pkg/disk_fallback_service.go", "package pkg\n")

	if app.ws.ByPath["pkg/disk_fallback_service.go"] != nil {
		t.Fatal("disk file unexpectedly already in workspace ByPath")
	}

	app.paletteQuery = "service"
	rows := app.paletteRows()

	if len(rows) < 3 {
		t.Fatalf("expected at least 3 rows, got %d: %+v", len(rows), rows)
	}

	// Verify in-memory matches come before disk fallback matches.
	foundDisk := -1
	for i, r := range rows {
		if r.Path == "pkg/disk_fallback_service.go" {
			foundDisk = i
			break
		}
	}

	if foundDisk < 0 {
		t.Fatalf("pkg/disk_fallback_service.go not found in palette rows: %+v", rows)
	}

	for i := 0; i < foundDisk; i++ {
		p := rows[i].Path
		if app.ws.ByPath[p] == nil {
			t.Fatalf("row at index %d (%s) is not an in-memory file but appeared before disk file at %d", i, p, foundDisk)
		}
	}
}

// TestPaletteFallbackDiscoversFilesOnDisk verifies that Quick Open finds files on
// disk that match the query even when they have not yet been loaded into a.ws.Files.
func TestPaletteFallbackDiscoversFilesOnDisk(t *testing.T) {
	dir := t.TempDir()
	writeFile(t, dir, "README.md", "# Test Project\n")
	writeFile(t, dir, "main.go", "package main\n")

	app, err := New(dir)
	if err != nil {
		t.Fatalf("New: %v", err)
	}

	// Create lazy file on disk after workspace load.
	writeFile(t, dir, "server/lazy_endpoint.go", "package server\n\nfunc Handle() string { return \"ok\" }\n")

	if app.ws.ByPath["server/lazy_endpoint.go"] != nil {
		t.Fatal("lazy_endpoint.go should not be in memory before palette search")
	}

	app.paletteQuery = "lazy"
	rows := app.paletteRows()

	if len(rows) != 1 {
		t.Fatalf("expected 1 row for query 'lazy', got %d: %+v", len(rows), rows)
	}

	row := rows[0]
	if row.Path != "server/lazy_endpoint.go" {
		t.Fatalf("unexpected path: got %s, want server/lazy_endpoint.go", row.Path)
	}
	if row.Name != "lazy_endpoint.go" {
		t.Fatalf("unexpected name: got %s, want lazy_endpoint.go", row.Name)
	}
	if row.Dir != "server" {
		t.Fatalf("unexpected dir: got %s, want server", row.Dir)
	}
	if row.Lang != "go" {
		t.Fatalf("unexpected lang: got %s, want go", row.Lang)
	}
	if row.Badge != "Go" {
		t.Fatalf("unexpected badge: got %s, want Go", row.Badge)
	}
}

// TestPaletteSelectFallbackResultOpensDocument verifies that selecting or opening
// a file discovered via disk fallback registers it in ByPath and Files, creates a
// tab, and renders syntax highlighting tokens.
func TestPaletteSelectFallbackResultOpensDocument(t *testing.T) {
	dir := t.TempDir()
	writeFile(t, dir, "README.md", "# Test Project\n")
	writeFile(t, dir, "main.go", "package main\n")

	app, err := New(dir)
	if err != nil {
		t.Fatalf("New: %v", err)
	}

	// Create lazy file on disk after initial scan.
	content := "package feature\n\n// TokenizedComment\nfunc Compute(x int) int {\n\treturn x * 2\n}\n"
	writeFile(t, dir, "pkg/feature/lazy_controller.go", content)

	if app.ws.ByPath["pkg/feature/lazy_controller.go"] != nil {
		t.Fatal("lazy_controller.go should not be in workspace ByPath before opening")
	}

	app.paletteOpen = true
	app.paletteQuery = "controller"
	rows := app.paletteRows()

	if len(rows) == 0 {
		t.Fatal("expected palette to find lazy_controller.go via disk fallback")
	}

	targetPath := rows[0].Path
	if targetPath != "pkg/feature/lazy_controller.go" {
		t.Fatalf("unexpected target path: %s", targetPath)
	}

	// Open the discovered file.
	app.openFile(targetPath)

	// 1. Verify ByPath and Files have the file registered.
	meta := app.ws.ByPath[targetPath]
	if meta == nil {
		t.Fatalf("%s not registered in a.ws.ByPath", targetPath)
	}
	if meta.Lang != "go" {
		t.Fatalf("expected lang 'go', got %q", meta.Lang)
	}

	foundInFiles := false
	for _, f := range app.ws.Files {
		if f.Path == targetPath {
			foundInFiles = true
			break
		}
	}
	if !foundInFiles {
		t.Fatalf("%s not found in a.ws.Files", targetPath)
	}

	// 2. Verify tab is open and active.
	if len(app.tabs) == 0 || app.tabs[app.active] != targetPath {
		t.Fatalf("expected active tab %q, got tabs=%v active=%d", targetPath, app.tabs, app.active)
	}

	// 3. Verify doc is loaded and syntax highlighting tokens are rendered.
	doc := app.docs[targetPath]
	if doc == nil {
		t.Fatalf("doc for %s not found in a.docs", targetPath)
	}
	if doc.Err != "" {
		t.Fatalf("doc error: %s", doc.Err)
	}
	if len(doc.Lines) == 0 {
		t.Fatal("doc lines are empty")
	}
	if len(doc.Tokens) == 0 {
		t.Fatal("doc tokens are empty (syntax highlighting failed)")
	}

	// 4. Verify palette state was reset.
	if app.paletteOpen {
		t.Fatal("expected palette to be closed after opening file")
	}
}

// TestPaletteGitFallback verifies that git ls-files is used when ripgrep is unavailable.
func TestPaletteGitFallback(t *testing.T) {
	if _, err := exec.LookPath("git"); err != nil {
		t.Skip("git not installed")
	}

	dir := t.TempDir()
	runGit(t, dir, "init", "-q", "-b", "main")
	runGit(t, dir, "config", "user.name", "Palette Test")
	runGit(t, dir, "config", "user.email", "palette@example.com")

	writeFile(t, dir, "README.md", "# Base\n")
	runGit(t, dir, "add", "-A")
	runGit(t, dir, "commit", "-q", "-m", "init")

	app, err := New(dir)
	if err != nil {
		t.Fatalf("New: %v", err)
	}

	// Add a new tracked file and an untracked file via git.
	writeFile(t, dir, "git_tracked.go", "package main\n")
	runGit(t, dir, "add", "git_tracked.go")
	writeFile(t, dir, "git_untracked.go", "package main\n")

	// Mock diskFileSearcher to test git directly.
	old := diskFileSearcher
	defer func() { diskFileSearcher = old }()

	diskFileSearcher = func(ctx context.Context, root string) ([]string, error) {
		out, err := gitBytes(ctx, root, "ls-files", "--cached", "--others", "--exclude-standard")
		if err != nil {
			return nil, err
		}
		return parseFileLines(out), nil
	}

	app.paletteQuery = "git_"
	rows := app.paletteRows()

	if len(rows) < 2 {
		t.Fatalf("expected at least 2 rows from git fallback, got %d: %+v", len(rows), rows)
	}
}

// TestPaletteMaxRowsCapsAtEight verifies that palette results never exceed paletteMaxRows.
func TestPaletteMaxRowsCapsAtEight(t *testing.T) {
	dir := t.TempDir()
	writeFile(t, dir, "README.md", "# Test\n")
	for i := 0; i < 15; i++ {
		writeFile(t, dir, "cmd/tool_"+string(rune('a'+i))+".go", "package main\n")
	}

	app, err := New(dir)
	if err != nil {
		t.Fatalf("New: %v", err)
	}

	app.paletteQuery = "tool"
	rows := app.paletteRows()

	if len(rows) != paletteMaxRows {
		t.Fatalf("expected exactly %d rows, got %d", paletteMaxRows, len(rows))
	}
}
