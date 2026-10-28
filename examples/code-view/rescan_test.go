package main

import (
	"os"
	"path/filepath"
	"slices"
	"testing"
)

// testWorkspace builds a workspace from explicit metas so a delta can be
// checked without touching disk.
func testWorkspace(metas ...*fileMeta) *workspace {
	ws := &workspace{ByPath: make(map[string]*fileMeta, len(metas))}
	for _, meta := range metas {
		ws.ByPath[meta.Path] = meta
		ws.Files = append(ws.Files, meta)
	}

	return ws
}

func TestDiffWorkspaceAddedRemovedChanged(t *testing.T) {
	t.Parallel()

	old := testWorkspace(
		&fileMeta{Path: "keep.go", Size: 10, Lang: langGo},
		&fileMeta{Path: "size.go", Size: 10, Lang: langGo},
		&fileMeta{Path: "lang.go", Size: 10, Lang: langGo},
		&fileMeta{Path: "gone.go", Size: 10, Lang: langGo},
		&fileMeta{Path: "z-old.txt", Size: 3, Lang: langPlain},
	)
	next := testWorkspace(
		&fileMeta{Path: "keep.go", Size: 10, Lang: langGo},
		&fileMeta{Path: "size.go", Size: 20, Lang: langGo},
		&fileMeta{Path: "lang.go", Size: 10, Lang: langJS},
		&fileMeta{Path: "z-new.txt", Size: 4, Lang: langPlain},
		&fileMeta{Path: "a-new.txt", Size: 4, Lang: langPlain},
	)

	delta := diffWorkspace(old, next)

	if want := []string{"a-new.txt", "z-new.txt"}; !slices.Equal(delta.Added, want) {
		t.Errorf("added = %v, want %v", delta.Added, want)
	}

	if want := []string{"gone.go", "z-old.txt"}; !slices.Equal(delta.Removed, want) {
		t.Errorf("removed = %v, want %v", delta.Removed, want)
	}

	if want := []string{"lang.go", "size.go"}; !slices.Equal(delta.Changed, want) {
		t.Errorf("changed = %v, want %v", delta.Changed, want)
	}

	if delta.Same != 1 {
		t.Errorf("same = %d, want 1", delta.Same)
	}

	same := diffWorkspace(old, old)
	if same.Same != len(old.Files) || len(same.Added)+len(same.Removed)+len(same.Changed) != 0 {
		t.Errorf("self delta = %+v", same)
	}
}

func TestDiffWorkspaceNil(t *testing.T) {
	t.Parallel()

	if delta := diffWorkspace(nil, nil); len(delta.Added)+len(delta.Removed)+len(delta.Changed) != 0 || delta.Same != 0 {
		t.Errorf("nil delta = %+v", delta)
	}

	next := testWorkspace(
		&fileMeta{Path: "b.go", Size: 2, Lang: langGo},
		&fileMeta{Path: "a.go", Size: 1, Lang: langGo},
	)

	added := diffWorkspace(nil, next)
	if want := []string{"a.go", "b.go"}; !slices.Equal(added.Added, want) {
		t.Errorf("added = %v, want %v", added.Added, want)
	}

	if len(added.Removed)+len(added.Changed) != 0 || added.Same != 0 {
		t.Errorf("added delta = %+v", added)
	}

	removed := diffWorkspace(next, nil)
	if want := []string{"a.go", "b.go"}; !slices.Equal(removed.Removed, want) {
		t.Errorf("removed = %v, want %v", removed.Removed, want)
	}

	if len(removed.Added)+len(removed.Changed) != 0 || removed.Same != 0 {
		t.Errorf("removed delta = %+v", removed)
	}

	// A workspace carrying only Files still compares by path.
	filesOnly := &workspace{Files: []*fileMeta{{Path: "c.go", Size: 3, Lang: langGo}}}
	if delta := diffWorkspace(nil, filesOnly); !slices.Equal(delta.Added, []string{"c.go"}) {
		t.Errorf("files-only added = %v", delta.Added)
	}
}

func TestRescanDiskWorkspace(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()

	write := func(name, body string) {
		t.Helper()

		if err := os.WriteFile(filepath.Join(dir, name), []byte(body), 0o600); err != nil {
			t.Fatal(err)
		}
	}

	write("keep.go", "package main\n")
	write("grow.go", "x")
	write("gone.go", "y")

	old, err := loadDiskWorkspace(dir)
	if err != nil {
		t.Fatal(err)
	}

	write("grow.go", "a longer body")
	write("fresh.txt", "new")

	if err := os.Remove(filepath.Join(dir, "gone.go")); err != nil {
		t.Fatal(err)
	}

	next, delta, err := rescanDiskWorkspace(dir, old, 100)
	if err != nil {
		t.Fatal(err)
	}

	if next == nil || next.Sample {
		t.Fatalf("next = %+v", next)
	}

	if want := []string{"fresh.txt"}; !slices.Equal(delta.Added, want) {
		t.Errorf("added = %v, want %v", delta.Added, want)
	}

	if want := []string{"gone.go"}; !slices.Equal(delta.Removed, want) {
		t.Errorf("removed = %v, want %v", delta.Removed, want)
	}

	if want := []string{"grow.go"}; !slices.Equal(delta.Changed, want) {
		t.Errorf("changed = %v, want %v", delta.Changed, want)
	}

	if delta.Same != 1 {
		t.Errorf("same = %d, want 1", delta.Same)
	}

	if next.ByPath["gone.go"] != nil || next.ByPath["fresh.txt"] == nil {
		t.Errorf("next files = %v", next.ByPath)
	}

	_, again, err := rescanDiskWorkspace(dir, next, 100)
	if err != nil {
		t.Fatal(err)
	}

	if len(again.Added)+len(again.Removed)+len(again.Changed) != 0 || again.Same != len(next.Files) {
		t.Errorf("unchanged rescan = %+v", again)
	}

	if _, _, err := rescanDiskWorkspace(filepath.Join(dir, "missing"), old, 100); err == nil {
		t.Error("rescan of a missing root succeeded")
	}
}

func TestGitMetadataChange(t *testing.T) {
	t.Parallel()

	cases := []struct {
		path string
		want bool
	}{
		{".git/HEAD", true},
		{".git/packed-refs", true},
		{".git/index", true},
		{".git/refs/heads/main", true},
		{".git/refs/heads/main.lock", true},
		{".git/refs/tags/v1.0.0", true},
		{".git/logs/HEAD", true},
		{".git/logs/refs/heads/main", true},
		{"/repo/.git/HEAD", true},
		{filepath.Join("repo", ".git", "refs", "heads", "main"), true},
		{".git/objects/ab/cdef0123", false},
		{".git/objects/pack/pack-abc.pack", false},
		{".git", false},
		{".git/refs", false},
		{".gitignore", false},
		{"notes/.gitignore", false},
		{".github/workflows/ci.yml", false},
		{"src/main.go", false},
	}

	for _, tc := range cases {
		if got := isGitMetadataChange(tc.path); got != tc.want {
			t.Errorf("isGitMetadataChange(%q) = %v, want %v", tc.path, got, tc.want)
		}
	}
}
