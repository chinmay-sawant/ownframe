package main

import (
	"slices"
	"testing"
)

// navChangeSet builds a change set from bare paths.
func navChangeSet(paths ...string) changeSet {
	files := make([]changedFile, 0, len(paths))

	for _, path := range paths {
		files = append(files, changedFile{Path: path})
	}

	return changeSet{Files: files}
}

func TestModifiedNavNextPrevWrap(t *testing.T) {
	t.Parallel()

	nav := newModifiedNav(navChangeSet(
		"web/theme.css",
		"backend/probe.go",
		"backend/store.go",
		"backend/probe.go",
	))

	want := []string{"backend/probe.go", "backend/store.go", "web/theme.css"}
	if !slices.Equal(nav.Files, want) {
		t.Fatalf("files = %v, want %v", nav.Files, want)
	}

	if nav.Len() != len(want) {
		t.Fatalf("len = %d, want %d", nav.Len(), len(want))
	}

	if path, ok := nav.Current(); !ok || path != want[0] {
		t.Fatalf("current = %q %v, want %q", path, ok, want[0])
	}

	for _, path := range want[1:] {
		if got, ok := nav.Next(); !ok || got != path {
			t.Fatalf("next = %q %v, want %q", got, ok, path)
		}
	}

	if got, ok := nav.Next(); !ok || got != want[0] {
		t.Fatalf("next wrap = %q %v, want %q", got, ok, want[0])
	}

	if got, ok := nav.Prev(); !ok || got != want[2] {
		t.Fatalf("prev wrap = %q %v, want %q", got, ok, want[2])
	}

	if got, ok := nav.Prev(); !ok || got != want[1] {
		t.Fatalf("prev = %q %v, want %q", got, ok, want[1])
	}
}

func TestModifiedNavSelectAndKeep(t *testing.T) {
	t.Parallel()

	nav := newModifiedNav(navChangeSet("a.go", "b.go", "c.go"))

	if !nav.Select("c.go") {
		t.Fatal("select c.go failed")
	}

	if got, ok := nav.Current(); !ok || got != "c.go" {
		t.Fatalf("current = %q %v, want c.go", got, ok)
	}

	if nav.Select("missing.go") {
		t.Fatal("select missing.go succeeded")
	}

	if got, ok := nav.Current(); !ok || got != "c.go" {
		t.Fatalf("current after miss = %q %v, want c.go", got, ok)
	}

	// The selected file survives, so the cursor follows it after a shift.
	nav.Keep([]string{"web.css", "c.go", "a.go"})
	if want := []string{"a.go", "c.go"}; !slices.Equal(nav.Files, want) {
		t.Fatalf("files after keep = %v, want %v", nav.Files, want)
	}

	if got, ok := nav.Current(); !ok || got != "c.go" {
		t.Fatalf("current after keep = %q %v, want c.go", got, ok)
	}

	// The selected file disappeared, so the cursor clamps to the last file.
	nav.Keep([]string{"a.go", "b.go"})
	if want := []string{"a.go"}; !slices.Equal(nav.Files, want) {
		t.Fatalf("files after clamp = %v, want %v", nav.Files, want)
	}

	if got, ok := nav.Current(); !ok || got != "a.go" {
		t.Fatalf("current after clamp = %q %v, want a.go", got, ok)
	}

	// Duplicate keep entries do not duplicate files.
	nav = newModifiedNav(navChangeSet("a.go", "b.go"))
	nav.Keep([]string{"b.go", "b.go"})
	if want := []string{"b.go"}; !slices.Equal(nav.Files, want) {
		t.Fatalf("files after duplicate keep = %v, want %v", nav.Files, want)
	}

	if nav.Index != 0 {
		t.Fatalf("index = %d, want 0", nav.Index)
	}
}

func TestModifiedNavEmpty(t *testing.T) {
	t.Parallel()

	nav := newModifiedNav(changeSet{})
	if nav.Len() != 0 || nav.Index != 0 {
		t.Fatalf("empty nav = %d files, index %d", nav.Len(), nav.Index)
	}

	if got, ok := nav.Current(); ok || got != "" {
		t.Fatalf("current = %q %v, want none", got, ok)
	}

	if got, ok := nav.Next(); ok || got != "" {
		t.Fatalf("next = %q %v, want none", got, ok)
	}

	if got, ok := nav.Prev(); ok || got != "" {
		t.Fatalf("prev = %q %v, want none", got, ok)
	}

	if nav.Label() != "" {
		t.Fatalf("label = %q, want empty", nav.Label())
	}

	if nav.Select("a.go") {
		t.Fatal("select on an empty nav succeeded")
	}

	nav.Keep([]string{"a.go"})
	if nav.Len() != 0 || nav.Index != 0 {
		t.Fatalf("keep on empty nav = %d files, index %d", nav.Len(), nav.Index)
	}

	nav.SetOpen(true)
	if !nav.Open {
		t.Fatal("open flag not set")
	}

	// Empty paths in the change set are skipped.
	blank := newModifiedNav(navChangeSet("", ""))
	if blank.Len() != 0 {
		t.Fatalf("blank nav len = %d, want 0", blank.Len())
	}
}

func TestModifiedNavLabel(t *testing.T) {
	t.Parallel()

	nav := newModifiedNav(sampleChangeSet())
	if nav.Len() != 7 {
		t.Fatalf("sample len = %d, want 7", nav.Len())
	}

	if !nav.Select("config/services.yaml") {
		t.Fatal("select config/services.yaml failed")
	}

	if got, want := nav.Label(), "3/7 config/services.yaml"; got != want {
		t.Fatalf("label = %q, want %q", got, want)
	}

	if !nav.Select("web/theme.css") {
		t.Fatal("select web/theme.css failed")
	}

	if got, want := nav.Label(), "7/7 web/theme.css"; got != want {
		t.Fatalf("last label = %q, want %q", got, want)
	}

	single := newModifiedNav(navChangeSet("app.go"))
	if got, want := single.Label(), "1/1 app.go"; got != want {
		t.Fatalf("single label = %q, want %q", got, want)
	}
}
