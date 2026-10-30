package main

import (
	"context"
	"strings"
	"testing"
)

func TestNavHistoryBackForward(t *testing.T) {
	dir := t.TempDir()
	writeFile(t, dir, "README.md", "# Fixture\n")
	writeFile(t, dir, "a.go", "package main\n")
	writeFile(t, dir, "b.go", "package main\n")

	app, err := New(dir)
	if err != nil {
		t.Fatal(err)
	}

	app.openFile("a.go")
	app.openFile("b.go")

	if !app.canNavBack() {
		t.Fatal("expected back history")
	}

	if app.View().Active != "b.go" {
		t.Fatalf("active = %q", app.View().Active)
	}

	if !app.goNavBack() {
		t.Fatal("go back failed")
	}

	if app.View().Active != "a.go" {
		t.Fatalf("after back active = %q", app.View().Active)
	}

	if !app.canNavForward() {
		t.Fatal("expected forward history")
	}

	if !app.goNavForward() {
		t.Fatal("go forward failed")
	}

	if app.View().Active != "b.go" {
		t.Fatalf("after forward active = %q", app.View().Active)
	}
}

func TestNavHistoryTruncatesForward(t *testing.T) {
	dir := t.TempDir()
	writeFile(t, dir, "README.md", "# Fixture\n")
	writeFile(t, dir, "a.go", "package main\n")
	writeFile(t, dir, "b.go", "package main\n")
	writeFile(t, dir, "c.go", "package main\n")

	app, err := New(dir)
	if err != nil {
		t.Fatal(err)
	}

	app.openFile("a.go")
	app.openFile("b.go")

	if !app.goNavBack() {
		t.Fatal("go back failed")
	}

	app.openFile("c.go")

	if app.canNavForward() {
		t.Fatal("forward history should be dropped after a new navigation")
	}

	if app.View().Active != "c.go" {
		t.Fatalf("active = %q", app.View().Active)
	}
}

func TestNavHistoryButtonsAndShortcut(t *testing.T) {
	dir := t.TempDir()
	writeFile(t, dir, "README.md", "# Fixture\n")
	writeFile(t, dir, "a.go", "package main\n")
	writeFile(t, dir, "b.go", "package main\n")

	app, err := New(dir)
	if err != nil {
		t.Fatal(err)
	}

	app.openFile("a.go")
	app.openFile("b.go")

	view := app.View()
	if !view.NavBack || view.NavForward {
		t.Fatalf("buttons back=%v forward=%v", view.NavBack, view.NavForward)
	}

	if !strings.Contains(view.NavBackTitle, "a.go") {
		t.Fatalf("back title = %q", view.NavBackTitle)
	}

	ctx := context.Background()
	if err := app.onKeyDown(ctx, "alt+arrowleft"); err != nil {
		t.Fatal(err)
	}

	if app.View().Active != "a.go" {
		t.Fatalf("alt+left active = %q", app.View().Active)
	}

	if err := app.onKeyDown(ctx, "alt+arrowright"); err != nil {
		t.Fatal(err)
	}

	if app.View().Active != "b.go" {
		t.Fatalf("alt+right active = %q", app.View().Active)
	}

	clickID(t, app, "nav-back")
	if app.View().Active != "a.go" {
		t.Fatalf("button back active = %q", app.View().Active)
	}

	clickID(t, app, "nav-forward")
	if app.View().Active != "b.go" {
		t.Fatalf("button forward active = %q", app.View().Active)
	}
}

func TestNavHistorySessionRoundTrip(t *testing.T) {
	dir := t.TempDir()
	writeFile(t, dir, "README.md", "# Fixture\n")
	writeFile(t, dir, "a.go", "package main\n")
	writeFile(t, dir, "b.go", "package main\n")

	app, err := New(dir)
	if err != nil {
		t.Fatal(err)
	}

	app.openFile("a.go")
	app.openFile("b.go")

	state := app.sessionState()
	if len(state.Nav) < 2 || state.NavAt != len(state.Nav)-1 {
		t.Fatalf("session nav = %+v", state.Nav)
	}

	restored, err := New(dir)
	if err != nil {
		t.Fatal(err)
	}

	restored.restoreState(state.sanitize(dir))

	if !restored.canNavBack() {
		t.Fatal("restored history has no back entry")
	}

	if got := restored.navBackPath(); got == "" {
		t.Fatal("restored back target is empty")
	}
}
