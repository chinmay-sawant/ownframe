package main

import (
	"context"
	"os/exec"
	"strings"
	"testing"

	"github.com/chinmay-sawant/ownframe"
)

func TestClipDiffPath(t *testing.T) {
	name := "a_very_long_component_filename_for_wrap_check.tsx"
	dir := "some/quite/long/nested/directory/structure/with_many_levels"

	if got, want := len([]rune(name)), 49; got != want {
		t.Fatalf("fixture name is %d runes, want %d", got, want)
	}

	short, dshort := clipDiffPath(886, 1280, "probe.go", "backend")
	if short != "probe.go" || dshort != "backend" {
		t.Fatalf("short path clipped to %q %q", short, dshort)
	}

	short, dshort = clipDiffPath(886, 1280, name, dir)
	if short != name {
		t.Fatalf("name clipped to %q, want full name", short)
	}

	if !strings.HasSuffix(dshort, "…") {
		t.Fatalf("dir %q has no ellipsis", dshort)
	}

	if got, want := len([]rune(dshort)), 83-49-2+1; got != want {
		t.Fatalf("dir is %d runes, want %d", got, want)
	}

	if _, dshort = clipDiffPath(406, 720, name, dir); dshort != "" {
		t.Fatalf("narrow dir = %q, want empty (hidden by CSS)", dshort)
	}

	if got, _ := clipDiffPath(100, 700, name, dir); !strings.HasSuffix(got, "…") {
		t.Fatalf("tiny budget name = %q, want ellipsis", got)
	}

	if got, dshort := clipDiffPath(886, 1280, "", ""); got != "" || dshort != "" {
		t.Fatalf("empty path = %q %q", got, dshort)
	}
}

func TestDiffPathStaysOnOneLine(t *testing.T) {
	if _, err := exec.LookPath("git"); err != nil {
		t.Skip("git is not installed")
	}

	dir := t.TempDir()
	runGit(t, dir, "init", "-q", "-b", "main")
	runGit(t, dir, "config", "user.name", "Code View Test")
	runGit(t, dir, "config", "user.email", "code-view@example.com")

	name := "a_very_long_component_filename_for_wrap_check.tsx"
	long := "some/quite/long/nested/directory/structure/with_many_levels/" + name
	writeFile(t, dir, long, "line1\nline2\nline3\n")
	writeFile(t, dir, "README.md", "# Base\n")
	runGit(t, dir, "add", "-A")
	runGit(t, dir, "commit", "-q", "-m", "base")
	writeFile(t, dir, long, "line1\nline2 changed\nline3\nline4\n")

	app, err := New(dir)
	if err != nil {
		t.Fatal(err)
	}

	ctx := context.Background()
	app.reloadChanges(ctx, false)

	if err := app.onClick(ctx, ownframe.Box{Action: "mode-changes"}); err != nil {
		t.Fatal(err)
	}

	app.Page().SetSize(1200, 800)

	if err := app.Page().Redraw(ctx); err != nil {
		t.Fatal(err)
	}

	view := app.View()
	if view.Diff.NameShort != name {
		t.Fatalf("NameShort = %q, want full name", view.Diff.NameShort)
	}

	if !strings.HasSuffix(view.Diff.DirShort, "…") {
		t.Fatalf("DirShort = %q, want ellipsis", view.Diff.DirShort)
	}

	box, ok := boxByID(app.Page().Boxes(), "diff-path")
	if !ok {
		t.Fatal("missing diff-path")
	}

	if box.H > 20 {
		t.Fatalf("diff-path is %.0f px tall, want one 20 px line", box.H)
	}
}
