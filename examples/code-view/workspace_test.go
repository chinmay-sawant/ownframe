package main

import (
	"os"
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
