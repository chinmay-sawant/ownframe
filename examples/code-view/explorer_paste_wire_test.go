package main

import (
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/chinmay-sawant/ownframe"
)

// pressCtrlV sends the control chord the window delivers for paste.
func pressCtrlV(t *testing.T, app *App) {
	t.Helper()

	ctx := context.Background()
	if err := app.onKeyDown(ctx, "control"); err != nil {
		t.Fatal(err)
	}

	if err := app.onKeyDown(ctx, "v"); err != nil {
		t.Fatal(err)
	}

	if err := app.onKeyUp(ctx, "control"); err != nil {
		t.Fatal(err)
	}
}

func TestPasteFileListCreatesCopy(t *testing.T) {
	ownframe.UseMemory(true)
	defer ownframe.UseMemory(false)
	clearStagedCopy()
	defer clearStagedCopy()

	dir := t.TempDir()
	writeFile(t, dir, "hello.txt", "hello\n")

	app, err := New(dir)
	if err != nil {
		t.Fatal(err)
	}

	srcDir := t.TempDir()
	writeFile(t, srcDir, "note.txt", "pasted bytes\n")

	src, err := filepath.Abs(filepath.Join(srcDir, "note.txt"))
	if err != nil {
		t.Fatal(err)
	}

	ownframe.Write("file://" + src + "\nfile:///does/not/exist.txt")
	pressCtrlV(t, app)

	raw, err := os.ReadFile(filepath.Join(dir, "note.txt")) //nolint:gosec // path is a t.TempDir fixture
	if err != nil {
		t.Fatalf("pasted file missing: %v", err)
	}

	if string(raw) != "pasted bytes\n" {
		t.Fatalf("pasted content = %q", raw)
	}

	if app.ws.ByPath["note.txt"] == nil {
		t.Fatal("pasted file is not registered in the workspace")
	}

	toast := app.View().Toast
	if !toast.Show || toast.Kind != "success" || !strings.Contains(toast.Message, "note.txt") {
		t.Fatalf("paste toast = %+v", toast)
	}
}

func TestPasteStagedCopyCreatesFile(t *testing.T) {
	ownframe.UseMemory(true)
	defer ownframe.UseMemory(false)
	clearStagedCopy()
	defer clearStagedCopy()

	dir := t.TempDir()
	writeFile(t, dir, "hello.txt", "hello\n")

	app, err := New(dir)
	if err != nil {
		t.Fatal(err)
	}

	app.stageExplorerCopy("hello.txt", false)

	// Unrelated clipboard text must not win over the staged copy.
	ownframe.Write("unrelated text")
	pressCtrlV(t, app)

	raw, err := os.ReadFile(filepath.Join(dir, "hello (copy).txt")) //nolint:gosec // path is a t.TempDir fixture
	if err != nil {
		t.Fatalf("staged paste missing: %v", err)
	}

	if string(raw) != "hello\n" {
		t.Fatalf("staged content = %q", raw)
	}

	toast := app.View().Toast
	if !toast.Show || toast.Kind != "success" || !strings.Contains(toast.Message, "hello (copy).txt") {
		t.Fatalf("staged toast = %+v", toast)
	}
}

func TestPastePlainTextCreatesFile(t *testing.T) {
	ownframe.UseMemory(true)
	defer ownframe.UseMemory(false)
	clearStagedCopy()
	defer clearStagedCopy()

	dir := t.TempDir()
	writeFile(t, dir, "hello.txt", "hello\n")

	app, err := New(dir)
	if err != nil {
		t.Fatal(err)
	}

	ownframe.Write("just some words")
	pressCtrlV(t, app)

	raw, err := os.ReadFile(filepath.Join(dir, "paste.txt")) //nolint:gosec // path is a t.TempDir fixture
	if err != nil {
		t.Fatalf("text paste missing: %v", err)
	}

	if string(raw) != "just some words" {
		t.Fatalf("text paste = %q", raw)
	}
}

func TestPasteWhileEditingTypesText(t *testing.T) {
	ownframe.UseMemory(true)
	defer ownframe.UseMemory(false)
	clearStagedCopy()
	defer clearStagedCopy()

	app := newTestApp(t)
	app.openFile("README.md")
	app.toggleEdit()

	if !app.editing {
		t.Fatal("editing did not start")
	}

	before := len(app.ws.Files)
	ctx := context.Background()

	if err := app.onPaste(ctx, "typed words"); err != nil {
		t.Fatal(err)
	}

	if !strings.Contains(app.edit.Text(), "typed words") {
		t.Fatalf("edit buffer = %q", app.edit.Text())
	}

	if len(app.ws.Files) != before {
		t.Fatal("explorer paste ran while editing")
	}

	if err := app.onKeyDown(ctx, "control"); err != nil {
		t.Fatal(err)
	}

	if err := app.onKeyDown(ctx, "v"); err != nil {
		t.Fatal(err)
	}

	if err := app.onKeyUp(ctx, "control"); err != nil {
		t.Fatal(err)
	}

	if len(app.ws.Files) != before {
		t.Fatal("Ctrl+V chord pasted a file while editing")
	}
}

func TestPasteSampleWorkspaceGuard(t *testing.T) {
	ownframe.UseMemory(true)
	defer ownframe.UseMemory(false)
	clearStagedCopy()
	defer clearStagedCopy()

	app := newTestApp(t)
	before := len(app.ws.Files)

	ownframe.Write("some text")
	pressCtrlV(t, app)

	if len(app.ws.Files) != before {
		t.Fatal("paste wrote into the sample workspace")
	}

	toast := app.View().Toast
	if !toast.Show || toast.Kind != "info" || !strings.Contains(toast.Title, "read-only") {
		t.Fatalf("sample toast = %+v", toast)
	}
}

func TestCtrlShiftVStillTogglesPreview(t *testing.T) {
	clearStagedCopy()
	defer clearStagedCopy()

	app := newTestApp(t)
	ctx := context.Background()

	if app.View().Lang != langMarkdown {
		t.Fatalf("initial language = %s", app.View().Lang)
	}

	if err := app.onKeyDown(ctx, "control"); err != nil {
		t.Fatal(err)
	}

	if err := app.onKeyDown(ctx, "shift"); err != nil {
		t.Fatal(err)
	}

	if err := app.onKeyDown(ctx, "v"); err != nil {
		t.Fatal(err)
	}

	if !app.View().Preview {
		t.Fatal("ctrl+shift+v did not turn preview on")
	}

	if err := app.onKeyUp(ctx, "control"); err != nil {
		t.Fatal(err)
	}

	if err := app.onKeyUp(ctx, "shift"); err != nil {
		t.Fatal(err)
	}
}

func TestPasteParsesClipboardFileList(t *testing.T) {
	srcDir := t.TempDir()
	writeFile(t, srcDir, "a.txt", "a\n")

	abs, err := filepath.Abs(filepath.Join(srcDir, "a.txt"))
	if err != nil {
		t.Fatal(err)
	}

	got := ParseClipboardFileList("file://" + abs + "\n# comment\nrelative.txt\n/does/not/exist.txt")

	if len(got) != 1 || got[0] != abs {
		t.Fatalf("parsed = %q, want [%q]", got, abs)
	}
}
