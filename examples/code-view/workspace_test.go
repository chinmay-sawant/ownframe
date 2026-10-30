package main

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

func TestSampleWorkspace(t *testing.T) {
	t.Parallel()

	ws, err := loadSampleWorkspace()
	if err != nil {
		t.Fatal(err)
	}

	if !ws.Sample {
		t.Fatal("workspace is not marked as the sample")
	}

	if len(ws.Files) != 17 {
		t.Fatalf("files = %d, want 17", len(ws.Files))
	}

	for _, path := range []string{"README.md", "backend/probe.go", "web/theme.css", "config/services.yaml"} {
		if ws.ByPath[path] == nil {
			t.Errorf("missing %s", path)
		}
	}

	backend := findChild(ws.Tree, "backend")
	if backend == nil || !backend.Dir || len(backend.Children) != 6 {
		t.Fatalf("backend = %+v", backend)
	}
}

// TestSampleHasNoJavaScript locks in the instruction that the shipped demo
// workspace carries no script files. JavaScript highlighting still ships and
// is covered by syntax_test.go.
func TestSampleHasNoJavaScript(t *testing.T) {
	t.Parallel()

	ws, err := loadSampleWorkspace()
	if err != nil {
		t.Fatal(err)
	}

	for _, file := range ws.Files {
		if file.Lang == langJS {
			t.Errorf("sample still ships a script file: %s", file.Path)
		}
	}
}

func TestOpenDocRejectsLongLine(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	path := filepath.Join(dir, "bundle.js")
	text := strings.Repeat("x", maxLineBytes+1) + "\n"

	if err := os.WriteFile(path, []byte(text), 0o600); err != nil {
		t.Fatal(err)
	}

	ws, err := loadDiskWorkspace(dir)
	if err != nil {
		t.Fatal(err)
	}

	doc := ws.openDoc("bundle.js")
	if doc.Err == "" {
		t.Fatal("long-line file opened")
	}

	if doc.Binary {
		t.Fatal("long-line file marked binary")
	}

	if len(doc.Lines) != 0 || len(doc.Tokens) != 0 {
		t.Fatalf("long-line file kept %d lines and %d token rows", len(doc.Lines), len(doc.Tokens))
	}
}

func TestOpenDoc(t *testing.T) {
	t.Parallel()

	ws, err := loadSampleWorkspace()
	if err != nil {
		t.Fatal(err)
	}

	doc := ws.openDoc("backend/probe.go")
	if doc.Err != "" {
		t.Fatalf("open probe.go: %s", doc.Err)
	}

	if doc.Lang != langGo {
		t.Fatalf("lang = %s", doc.Lang)
	}

	if len(doc.Lines) < 100 || len(doc.Tokens) != len(doc.Lines) {
		t.Fatalf("lines = %d tokens = %d", len(doc.Lines), len(doc.Tokens))
	}

	if len(doc.Outline) == 0 {
		t.Fatal("no outline for probe.go")
	}

	readme := ws.openDoc("README.md")
	if readme.Err != "" || !strings.Contains(strings.Join(readme.Lines, "\n"), "# Beacon") {
		t.Fatalf("README = %v", readme.Err)
	}
}

func TestDiskWorkspaceAndBinary(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()

	if err := os.WriteFile(filepath.Join(dir, "main.go"), []byte("package main\n\nfunc main() {}\n"), 0o600); err != nil {
		t.Fatal(err)
	}

	if err := os.WriteFile(filepath.Join(dir, "blob.bin"), append([]byte("text"), 0), 0o600); err != nil {
		t.Fatal(err)
	}

	if err := os.MkdirAll(filepath.Join(dir, "node_modules", "x"), 0o700); err != nil {
		t.Fatal(err)
	}

	if err := os.WriteFile(filepath.Join(dir, "node_modules", "x", "skip.js"), []byte("skip"), 0o600); err != nil {
		t.Fatal(err)
	}

	ws, err := loadDiskWorkspace(dir)
	if err != nil {
		t.Fatal(err)
	}

	if ws.Sample || ws.ByPath["main.go"] == nil || ws.ByPath["blob.bin"] == nil {
		t.Fatalf("files = %v", ws.ByPath)
	}

	if ws.ByPath["node_modules/x/skip.js"] != nil {
		t.Fatal("node_modules was scanned")
	}

	blob := ws.openDoc("blob.bin")
	if !blob.Binary {
		t.Fatalf("blob not binary: %+v", blob)
	}

	if doc := ws.openDoc("main.go"); doc.Err != "" || len(doc.Lines) != 3 {
		t.Fatalf("main.go = %+v", doc)
	}
}

// writeScanFiles fills dir with count text files the scanner accepts.
func writeScanFiles(t *testing.T, dir string, count int) {
	t.Helper()

	for i := range count {
		name := fmt.Sprintf("file%02d.txt", i)
		if err := os.WriteFile(filepath.Join(dir, name), []byte("package main\n"), 0o600); err != nil {
			t.Fatal(err)
		}
	}
}

func TestLoadDiskWorkspaceLimitTruncates(t *testing.T) {
	t.Parallel()

	for _, limit := range []int{1, 3} {
		t.Run(fmt.Sprintf("limit%d", limit), func(t *testing.T) {
			t.Parallel()

			dir := t.TempDir()
			writeScanFiles(t, dir, 5)

			ws, err := loadDiskWorkspaceWithLimit(dir, limit)
			if err != nil {
				t.Fatal(err)
			}

			if len(ws.Files) != limit {
				t.Fatalf("files = %d, want %d", len(ws.Files), limit)
			}

			if !ws.Truncated {
				t.Fatal("workspace not marked truncated")
			}

			if ws.Limit != limit {
				t.Fatalf("limit = %d, want %d", ws.Limit, limit)
			}

			if want := fmt.Sprintf("%d files", limit); !strings.Contains(ws.TruncatedNote, want) {
				t.Errorf("note %q does not contain %q", ws.TruncatedNote, want)
			}

			for _, word := range []string{"explorer", "quick open", "search", "raise the limit"} {
				if !strings.Contains(ws.TruncatedNote, word) {
					t.Errorf("note %q does not mention %q", ws.TruncatedNote, word)
				}
			}

			budget := ws.ScanBudget()
			if budget == "" || !strings.Contains(budget, fmt.Sprint(limit)) {
				t.Errorf("scan budget = %q, want the %d-file budget", budget, limit)
			}

			if ws.ScanMillis < 0 {
				t.Errorf("scan millis = %d", ws.ScanMillis)
			}
		})
	}
}

func TestLoadDiskWorkspaceExactLimitNotTruncated(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	writeScanFiles(t, dir, 3)

	ws, err := loadDiskWorkspaceWithLimit(dir, 3)
	if err != nil {
		t.Fatal(err)
	}

	if len(ws.Files) != 3 {
		t.Fatalf("files = %d, want 3", len(ws.Files))
	}

	if ws.Truncated {
		t.Fatal("exact-limit scan marked truncated")
	}

	if ws.Limit != 3 {
		t.Fatalf("limit = %d, want 3", ws.Limit)
	}

	if ws.TruncatedNote != "" {
		t.Fatalf("note = %q, want empty", ws.TruncatedNote)
	}

	if budget := ws.ScanBudget(); budget != "" {
		t.Fatalf("scan budget = %q, want empty", budget)
	}
}

func TestLoadDiskWorkspaceDefaultLimitNoTruncation(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	writeScanFiles(t, dir, 5)

	ws, err := loadDiskWorkspace(dir)
	if err != nil {
		t.Fatal(err)
	}

	if len(ws.Files) != 5 || ws.Truncated {
		t.Fatalf("files = %d truncated = %v", len(ws.Files), ws.Truncated)
	}

	if ws.Limit != maxWorkspaceFiles {
		t.Fatalf("limit = %d, want %d", ws.Limit, maxWorkspaceFiles)
	}

	if ws.TruncatedNote != "" || ws.ScanBudget() != "" {
		t.Fatalf("note = %q budget = %q", ws.TruncatedNote, ws.ScanBudget())
	}

	zero, err := loadDiskWorkspaceWithLimit(dir, 0)
	if err != nil {
		t.Fatal(err)
	}

	if zero.Limit != maxWorkspaceFiles || zero.Truncated || len(zero.Files) != 5 {
		t.Fatalf("zero limit = %d files = %d truncated = %v", zero.Limit, len(zero.Files), zero.Truncated)
	}
}

func TestSampleWorkspaceBudgetEmpty(t *testing.T) {
	t.Parallel()

	ws, err := loadSampleWorkspace()
	if err != nil {
		t.Fatal(err)
	}

	if ws.Truncated {
		t.Fatal("sample workspace marked truncated")
	}

	if ws.TruncatedNote != "" {
		t.Fatalf("sample note = %q, want empty", ws.TruncatedNote)
	}

	if budget := ws.ScanBudget(); budget != "" {
		t.Fatalf("sample budget = %q, want empty", budget)
	}

	if ws.Limit != 0 && ws.Limit != maxWorkspaceFiles {
		t.Fatalf("sample limit = %d", ws.Limit)
	}

	if ws.ScanMillis < 0 {
		t.Fatalf("sample scan millis = %d", ws.ScanMillis)
	}
}

func TestWorkspaceMarksGitIgnoredPaths(t *testing.T) {
	if _, err := exec.LookPath("git"); err != nil {
		t.Skip("git is not installed")
	}

	dir := t.TempDir()
	runGit(t, dir, "init", "-q")
	writeFile(t, dir, ".gitignore", "ignored/\n*.log\nsecret.txt\nnode_modules/\nempty-ignored/\n")
	writeFile(t, dir, "keep.go", "package keep\n")
	writeFile(t, dir, "secret.txt", "secret\n")
	writeFile(t, dir, "logs/app.log", "log\n")
	writeFile(t, dir, "ignored/a.go", "package ignored\n")
	writeFile(t, dir, "node_modules/pkg/index.js", "x\n")

	if err := os.MkdirAll(filepath.Join(dir, "empty-ignored"), 0o700); err != nil {
		t.Fatal(err)
	}

	ws, err := loadDiskWorkspace(dir)
	if err != nil {
		t.Fatal(err)
	}

	for _, path := range []string{"secret.txt", "logs/app.log", "ignored/a.go", "ignored"} {
		if !pathIgnored(map[string]bool{"secret.txt": true, "logs": true, "ignored": true}, path) {
			t.Errorf("%s not ignored by the path walk", path)
		}
	}

	if !ws.ByPath["secret.txt"].Ignored || !ws.ByPath["ignored/a.go"].Ignored {
		t.Fatal("workspace metadata did not mark ignored files")
	}

	if ws.ByPath["node_modules/pkg/index.js"] == nil {
		t.Fatal("scanner did not walk the ignored node_modules directory")
	}

	if !ws.ByPath["node_modules/pkg/index.js"].Ignored {
		t.Fatal("node_modules file was not marked ignored")
	}

	if node := findTreeNode(ws.Tree, "node_modules"); node == nil || !node.Ignored {
		t.Fatalf("ignored directory = %+v", node)
	} else if len(node.Children) == 0 {
		t.Fatalf("ignored directory has no children: %+v", node)
	}

	if node := findTreeNode(ws.Tree, "empty-ignored"); node == nil || !node.Ignored {
		t.Fatalf("empty ignored directory placeholder = %+v", node)
	}

	if ws.ByPath["keep.go"].Ignored {
		t.Fatal("keep.go was marked ignored")
	}

	var ignoredDir *fileNode

	for _, node := range ws.Tree {
		if node.Path == "ignored" {
			ignoredDir = node
		}
	}

	if ignoredDir == nil || !ignoredDir.Ignored {
		t.Fatalf("ignored directory = %+v", ignoredDir)
	}

	app, err := New(dir)
	if err != nil {
		t.Fatal(err)
	}

	if err := app.Page().Redraw(context.Background()); err != nil {
		t.Fatal(err)
	}

	found := false
	foundIgnored := false

	for _, row := range app.View().Tree {
		if row.Path == "keep.go" {
			found = true
			if row.Ignored {
				t.Fatal("keep.go row is ignored")
			}
		}

		if row.Path == "ignored" {
			foundIgnored = true
			if !row.Ignored {
				t.Fatal("ignored directory row is not marked")
			}
		}
	}

	if !found || !foundIgnored {
		t.Fatalf("tree rows keep=%v ignored=%v", found, foundIgnored)
	}

}

func TestLargeFileIsRefused(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	big := strings.Repeat("x", maxFileBytes+1)

	if err := os.WriteFile(filepath.Join(dir, "big.txt"), []byte(big), 0o600); err != nil {
		t.Fatal(err)
	}

	ws, err := loadDiskWorkspace(dir)
	if err != nil {
		t.Fatal(err)
	}

	if !ws.ByPath["big.txt"].Big {
		t.Fatal("big file not marked")
	}

	doc := ws.openDoc("big.txt")
	if doc.Err == "" || !strings.Contains(doc.Err, "larger than") {
		t.Fatalf("doc = %+v", doc)
	}
}

func TestHumanSize(t *testing.T) {
	t.Parallel()

	cases := map[int64]string{
		12:         "12 B",
		byteKB:     "1.0 KB",
		4 * byteKB: "4.0 KB",
		byteMB:     "1.0 MB",
	}
	for size, want := range cases {
		if got := humanSize(size); got != want {
			t.Errorf("humanSize(%d) = %q, want %q", size, got, want)
		}
	}
}
