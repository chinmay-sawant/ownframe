package main

import (
	"context"
	"os"
	"strings"
	"testing"
)

func TestBranchChangesCombineCommitsWithoutWorkingTree(t *testing.T) {
	dir, _ := historyRepo(t)
	runGit(t, dir, "switch", "-q", "-c", "feature/combined")
	writeFile(t, dir, "first.txt", "first committed change\n")
	runGit(t, dir, "add", "first.txt")
	runGit(t, dir, "commit", "-q", "-m", "first change")
	writeFile(t, dir, "second.txt", "second committed change\n")
	writeFile(t, dir, "first.txt", "first committed change\nupdated later\n")
	runGit(t, dir, "add", "first.txt", "second.txt")
	runGit(t, dir, "commit", "-q", "-m", "second change")
	runGit(t, dir, "switch", "-q", "main")

	base, err := defaultBranchRef(context.Background(), dir)
	if err != nil || base != "refs/heads/main" {
		t.Fatalf("default branch = %q, %v", base, err)
	}
	changes, err := loadBranchChanges(context.Background(), dir, base, "refs/heads/feature/combined")
	if err != nil {
		t.Fatal(err)
	}
	if len(changes.Files) != 2 || findFile(changes.Files, "first.txt") == nil || findFile(changes.Files, "second.txt") == nil {
		t.Fatalf("branch changes = %+v", changes.Files)
	}
	if findFile(changes.Files, "README.md") != nil {
		t.Fatal("uncommitted working-tree change leaked into branch comparison")
	}
	if first := findFile(changes.Files, "first.txt"); first.Add != 2 {
		t.Fatalf("combined first file: %+v", first)
	}
	app, err := New(dir)
	if err != nil {
		t.Fatal(err)
	}
	app.reloadChanges(context.Background(), false)
	app.setMode(context.Background(), "changes")
	app.pickBranch(context.Background(), "refs/heads/feature/combined")
	if err := app.Page().Redraw(context.Background()); err != nil {
		t.Fatal(err)
	}
	clickID(t, app, "history-branch-changes")
	view := app.View()
	if view.Diff.Source != "branch" || view.Diff.Count != 2 || view.FileNav.Total != 2 {
		t.Fatalf("branch view = %+v navigator = %+v", view.Diff, view.FileNav)
	}
	if !strings.Contains(view.Diff.Note, "feature/combined") || !strings.Contains(view.Diff.Note, "main") {
		t.Fatalf("comparison description = %q", view.Diff.Note)
	}
	app.fileNavStep(1)
	if view = app.View(); view.Diff.Source != "branch" || view.Diff.Path != "second.txt" {
		t.Fatalf("branch navigator = %+v", view.Diff)
	}
	runGit(t, dir, "switch", "-q", "feature/combined")
	writeFile(t, dir, "third.txt", "third change\n")
	runGit(t, dir, "add", "third.txt")
	runGit(t, dir, "commit", "-q", "-m", "third change")
	runGit(t, dir, "switch", "-q", "main")
	app.applyGitRefresh(context.Background(), watchUpdate{})
	if view = app.View(); view.Diff.Source != "branch" || view.Diff.Count != 3 || view.Diff.Path != "second.txt" {
		t.Fatalf("refreshed branch view = %+v", view.Diff)
	}
	app.showWorkingDiff()
	if view = app.View(); view.Diff.Source != "working" || view.Diff.Count != 1 || view.Diff.Files[0].Path != "README.md" {
		t.Fatalf("working tree = %+v", view.Diff)
	}
	if branch := runGit(t, dir, "branch", "--show-current"); branch != "main" {
		t.Fatalf("checkout changed to %s", branch)
	}
}

func TestBranchChangesSample(t *testing.T) {
	app := newTestApp(t)
	app.setMode(context.Background(), "changes")
	if err := app.Page().Redraw(context.Background()); err != nil {
		t.Fatal(err)
	}
	clickID(t, app, "history-branch-changes")
	if view := app.View(); view.Diff.Source != "branch" || view.Diff.Count != len(app.changes.Files) {
		t.Fatalf("sample branch diff = %+v", view.Diff)
	}
	if os.Getenv("OWNFRAME_CODE_VIEW_SCREENSHOT") == "1" {
		writeShot(t, app, "branch-changes")
	}
}

func TestDefaultBranchRequiresUnambiguousBase(t *testing.T) {
	dir, _ := historyRepo(t)
	runGit(t, dir, "branch", "master")
	if _, err := defaultBranchRef(context.Background(), dir); err == nil {
		t.Fatal("accepted ambiguous main and master branches")
	}
	runGit(t, dir, "update-ref", "refs/remotes/origin/main", "HEAD")
	runGit(t, dir, "symbolic-ref", "refs/remotes/origin/HEAD", "refs/remotes/origin/main")
	base, err := defaultBranchRef(context.Background(), dir)
	if err != nil || base != "refs/remotes/origin/main" {
		t.Fatalf("remote default = %q, %v", base, err)
	}
}
