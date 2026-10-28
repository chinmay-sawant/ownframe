package main

import (
	"fmt"
	"testing"
)

func TestEditorGroupOpenDedupes(t *testing.T) {
	t.Parallel()

	var group editorGroup

	index, evicted := group.Open("README.md", maxTabs)
	if index != 0 || evicted != "" || group.Len() != 1 || group.Path != "README.md" {
		t.Fatalf("first open: index=%d evicted=%q len=%d path=%q", index, evicted, group.Len(), group.Path)
	}

	if index, evicted = group.Open("backend/probe.go", maxTabs); index != 1 || evicted != "" {
		t.Fatalf("second open: index=%d evicted=%q", index, evicted)
	}

	if index, evicted = group.Open("README.md", maxTabs); index != 0 || evicted != "" {
		t.Fatalf("reopen: index=%d evicted=%q", index, evicted)
	}

	if group.Len() != 2 {
		t.Fatalf("reopen len = %d, want 2", group.Len())
	}

	if group.ActivePath() != "README.md" || group.Active != 0 {
		t.Fatalf("reopen active = %d %q, want 0 README.md", group.Active, group.ActivePath())
	}

	if index, evicted = group.Open("", maxTabs); index != -1 || evicted != "" || group.Len() != 2 {
		t.Fatalf("blank open: index=%d evicted=%q len=%d", index, evicted, group.Len())
	}
}

func TestEditorGroupEvictsOldest(t *testing.T) {
	t.Parallel()

	var group editorGroup

	for i := 0; i < maxTabs; i++ {
		index, evicted := group.Open(fmt.Sprintf("file%d.go", i), maxTabs)
		if index != i || evicted != "" {
			t.Fatalf("open %d: index=%d evicted=%q", i, index, evicted)
		}
	}

	index, evicted := group.Open("newest.go", maxTabs)
	if index != maxTabs-1 || evicted != "file0.go" {
		t.Fatalf("overflow: index=%d evicted=%q, want %d file0.go", index, evicted, maxTabs-1)
	}

	if group.Len() != maxTabs || group.Path != "newest.go" {
		t.Fatalf("overflow: len=%d path=%q", group.Len(), group.Path)
	}

	paths := group.Paths()
	if paths[0] != "file1.go" || paths[maxTabs-1] != "newest.go" {
		t.Fatalf("overflow paths = %v", paths)
	}

	// Reopening a tab that is already open must not evict when full.
	if index, evicted = group.Open("file1.go", maxTabs); index != 0 || evicted != "" || group.Len() != maxTabs {
		t.Fatalf("full reopen: index=%d evicted=%q len=%d", index, evicted, group.Len())
	}

	// A max below 1 falls back to maxTabs.
	var fallback editorGroup

	for i := 0; i <= maxTabs; i++ {
		fallback.Open(fmt.Sprintf("f%d.go", i), 0)
	}

	if fallback.Len() != maxTabs {
		t.Fatalf("fallback len = %d, want %d", fallback.Len(), maxTabs)
	}

	if fallback.Paths()[0] != "f1.go" {
		t.Fatalf("fallback first = %q, want f1.go", fallback.Paths()[0])
	}
}

func TestEditorGroupCloseKeepsActiveValid(t *testing.T) {
	t.Parallel()

	var group editorGroup

	for _, path := range []string{"a.go", "b.go", "c.go", "d.go"} {
		group.Open(path, maxTabs)
	}

	if !group.Select(2) {
		t.Fatal("select c.go failed")
	}

	closed, ok := group.Close(0)
	if !ok || closed != "a.go" {
		t.Fatalf("close a.go = %q %v", closed, ok)
	}

	if group.Active != 1 || group.Path != "c.go" {
		t.Fatalf("after close before active: active=%d path=%q", group.Active, group.Path)
	}

	closed, ok = group.Close(2)
	if !ok || closed != "d.go" || group.Active != 1 || group.Path != "c.go" {
		t.Fatalf("close after active: closed=%q ok=%v active=%d path=%q", closed, ok, group.Active, group.Path)
	}

	closed, ok = group.Close(1)
	if !ok || closed != "c.go" || group.Active != 0 || group.Path != "b.go" {
		t.Fatalf("close active: closed=%q ok=%v active=%d path=%q", closed, ok, group.Active, group.Path)
	}

	closed, ok = group.Close(0)
	if !ok || closed != "b.go" || group.Len() != 0 || group.Active != 0 || group.Path != "" {
		t.Fatalf("close last: closed=%q ok=%v len=%d active=%d path=%q", closed, ok, group.Len(), group.Active, group.Path)
	}

	if closed, ok = group.Close(0); ok || closed != "" {
		t.Fatalf("close empty = %q %v", closed, ok)
	}
}

func TestEditorGroupSelectAndReset(t *testing.T) {
	t.Parallel()

	var group editorGroup

	if group.Select(0) || group.Select(-1) || group.SelectPath("README.md") {
		t.Fatal("select on empty group reported true")
	}

	group.Open("a.go", maxTabs)
	group.Open("b.go", maxTabs)

	if group.Select(2) || !group.Select(0) || group.Path != "a.go" {
		t.Fatalf("select: active=%d path=%q", group.Active, group.Path)
	}

	if !group.SelectPath("b.go") || group.ActivePath() != "b.go" {
		t.Fatalf("select path: active=%d path=%q", group.Active, group.Path)
	}

	if group.SelectPath("missing.go") || group.ActivePath() != "b.go" {
		t.Fatalf("select missing: active=%d path=%q", group.Active, group.Path)
	}

	paths := []string{"README.md", " ", "README.md", "app.go", "", "view.go"}
	group.Reset(paths, 1)

	if group.Len() != 3 {
		t.Fatalf("reset len = %d, want 3", group.Len())
	}

	if group.Path != "app.go" || group.Active != 1 {
		t.Fatalf("reset active = %d %q, want 1 app.go", group.Active, group.Path)
	}

	if got := group.Paths(); got[0] != "README.md" || got[1] != "app.go" || got[2] != "view.go" {
		t.Fatalf("reset paths = %v", got)
	}

	paths[0] = "changed.go"

	if group.Paths()[0] != "README.md" {
		t.Fatalf("reset aliased input: %v", group.Paths())
	}

	group.Reset([]string{"a.go", " b.go "}, 9)

	if group.Active != 1 || group.Path != "b.go" || group.Paths()[0] != "a.go" {
		t.Fatalf("reset clamp: active=%d path=%q paths=%v", group.Active, group.Path, group.Paths())
	}

	many := make([]string, 0, maxTabs+2)

	for i := 0; i < maxTabs+2; i++ {
		many = append(many, fmt.Sprintf("m%d.go", i))
	}

	group.Reset(many, 0)

	if group.Len() != maxTabs {
		t.Fatalf("reset cap len = %d, want %d", group.Len(), maxTabs)
	}

	group.Reset(nil, 0)

	if group.Len() != 0 || group.Active != 0 || group.Path != "" || group.ActivePath() != "" {
		t.Fatalf("reset empty: len=%d active=%d path=%q", group.Len(), group.Active, group.Path)
	}
}

func TestEditorGroupClosePath(t *testing.T) {
	t.Parallel()

	var group editorGroup

	for _, path := range []string{"a.go", "b.go", "c.go"} {
		group.Open(path, maxTabs)
	}

	if !group.SelectPath("c.go") {
		t.Fatal("select c.go failed")
	}

	if !group.ClosePath("a.go") {
		t.Fatal("close path a.go reported false")
	}

	if group.Len() != 2 || group.ActivePath() != "c.go" || group.Path != "c.go" {
		t.Fatalf("after close a.go: len=%d active=%d path=%q", group.Len(), group.Active, group.Path)
	}

	if group.ClosePath("missing.go") || group.Len() != 2 {
		t.Fatalf("close missing: len=%d", group.Len())
	}

	if !group.ClosePath("c.go") {
		t.Fatal("close path c.go reported false")
	}

	if group.ActivePath() != "b.go" || group.Path != "b.go" {
		t.Fatalf("after close active: active=%d path=%q", group.Active, group.Path)
	}

	if !group.ClosePath("b.go") || group.Len() != 0 || group.Path != "" {
		t.Fatalf("after close last: len=%d path=%q", group.Len(), group.Path)
	}
}

func TestEditorGroupEmpty(t *testing.T) {
	t.Parallel()

	var group editorGroup

	if group.Len() != 0 || group.ActivePath() != "" || group.Paths() != nil {
		t.Fatalf("zero value: len=%d path=%q paths=%v", group.Len(), group.ActivePath(), group.Paths())
	}

	if group.ClosePath("a.go") {
		t.Fatal("close path on empty group reported true")
	}

	if closed, ok := group.Close(0); ok || closed != "" {
		t.Fatalf("close on empty group = %q %v", closed, ok)
	}

	group.Open("a.go", maxTabs)

	paths := group.Paths()
	paths[0] = "changed.go"

	if group.ActivePath() != "a.go" {
		t.Fatalf("paths aliased the group: %q", group.ActivePath())
	}
}
