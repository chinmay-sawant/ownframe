package main

import (
	"bytes"
	"os"
	"path/filepath"
	"testing"
)

func TestWorkspaceWriteFile(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	writeFile(t, dir, "main.go", "package main\n\nfunc main() {}\n")
	writeFile(t, dir, "pkg/util.go", "package pkg\n")

	ws, err := loadDiskWorkspace(dir)
	if err != nil {
		t.Fatal(err)
	}

	if !ws.writable() {
		t.Fatal("disk workspace is not writable")
	}

	edited := []byte("package main\n\n// edited by the viewer\n")
	if err := ws.writeFile("main.go", edited); err != nil {
		t.Fatal(err)
	}

	got, err := os.ReadFile(filepath.Join(dir, "main.go")) //nolint:gosec // path is a t.TempDir fixture
	if err != nil {
		t.Fatal(err)
	}

	if !bytes.Equal(got, edited) {
		t.Fatalf("main.go = %q, want %q", got, edited)
	}

	meta := ws.ByPath["main.go"]
	if meta == nil {
		t.Fatal("main.go missing from ByPath")
	}

	if meta.Size != int64(len(edited)) {
		t.Fatalf("size = %d, want %d", meta.Size, len(edited))
	}

	if meta.Big {
		t.Fatal("small file marked big")
	}

	nested := []byte("package pkg\n\nconst Edited = true\n")
	if err := ws.writeFile("pkg/util.go", nested); err != nil {
		t.Fatal(err)
	}

	got, err = os.ReadFile(filepath.Join(dir, "pkg", "util.go")) //nolint:gosec // path is a t.TempDir fixture
	if err != nil {
		t.Fatal(err)
	}

	if !bytes.Equal(got, nested) {
		t.Fatalf("pkg/util.go = %q, want %q", got, nested)
	}

	if size := ws.ByPath["pkg/util.go"].Size; size != int64(len(nested)) {
		t.Fatalf("nested size = %d, want %d", size, len(nested))
	}

	big := bytes.Repeat([]byte("x"), maxFileBytes+1)
	if err := ws.writeFile("main.go", big); err != nil {
		t.Fatal(err)
	}

	if !ws.ByPath["main.go"].Big {
		t.Fatal("large file not marked big")
	}

	if size := ws.ByPath["main.go"].Size; size != int64(len(big)) {
		t.Fatalf("large size = %d, want %d", size, len(big))
	}
}

func TestWorkspaceWriteRejectsEscapeAndUnknownPath(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	writeFile(t, dir, "main.go", "package main\n")

	ws, err := loadDiskWorkspace(dir)
	if err != nil {
		t.Fatal(err)
	}

	escapeName := filepath.Base(dir) + "-escape.txt"

	paths := []string{
		"../" + escapeName,
		"pkg/../../" + escapeName,
		filepath.Join(dir, "abs.go"),
		"missing.go",
	}

	for _, path := range paths {
		if err := ws.writeFile(path, []byte("nope\n")); err == nil {
			t.Errorf("writeFile(%q) succeeded", path)
		}
	}

	if _, err := os.Stat(filepath.Join(dir, "..", escapeName)); !os.IsNotExist(err) {
		t.Fatalf("escape file exists: %v", err)
	}

	if _, err := os.Stat(filepath.Join(dir, "abs.go")); !os.IsNotExist(err) {
		t.Fatalf("absolute write created a file: %v", err)
	}

	if _, err := os.Stat(filepath.Join(dir, "missing.go")); !os.IsNotExist(err) {
		t.Fatalf("unknown write created a file: %v", err)
	}
}

func TestSampleWorkspaceReadOnly(t *testing.T) {
	t.Parallel()

	ws, err := loadSampleWorkspace()
	if err != nil {
		t.Fatal(err)
	}

	if ws.writable() {
		t.Fatal("sample workspace is writable")
	}

	for _, path := range []string{"README.md", "../outside.txt", "missing.go"} {
		err := ws.writeFile(path, []byte("rewritten\n"))
		if err == nil {
			t.Fatalf("writeFile(%q) succeeded", path)
		}

		if err.Error() != sampleReadOnlyMessage {
			t.Fatalf("writeFile(%q) error = %q, want %q", path, err.Error(), sampleReadOnlyMessage)
		}
	}
}
