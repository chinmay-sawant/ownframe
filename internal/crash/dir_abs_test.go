package crash_test

import (
	"path/filepath"
	"testing"

	"github.com/chinmay-sawant/ownframe/internal/crash"
)

func TestSetDirStoresAbsolute(t *testing.T) {
	base := t.TempDir()
	t.Cleanup(func() { _ = crash.SetDir(base) })
	if err := crash.SetDir("crash-rel-dir"); err != nil {
		t.Fatal(err)
	}
	got := crash.Dir()
	if !filepath.IsAbs(got) {
		t.Fatalf("Dir = %q, want absolute", got)
	}
	if err := crash.SetDir("../escape-crash"); err == nil {
		t.Fatal("traversal accepted")
	}
	abs := filepath.Join(base, "reports")
	if err := crash.SetDir(abs); err != nil {
		t.Fatal(err)
	}
	if got := crash.Dir(); got != abs {
		t.Fatalf("Dir = %q, want %q", got, abs)
	}
}

func TestSetDirRejectsTraversal(t *testing.T) {
	base := t.TempDir()
	t.Cleanup(func() { _ = crash.SetDir(base) })

	for _, dir := range []string{
		"/allowed/../secret",
		base + "/../escape",
		"..",
		"bad\x00dir",
		"bad\ndir",
	} {
		if err := crash.SetDir(dir); err == nil {
			t.Fatalf("SetDir(%q) accepted", dir)
		}
	}
}
