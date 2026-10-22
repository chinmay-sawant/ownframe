package crash_test

import (
	"path/filepath"
	"strings"
	"testing"

	"github.com/chinmay-sawant/ownframe/internal/crash"
)

func fuzzSetDirOnce(t *testing.T, dir string) {
	t.Helper()
	base := t.TempDir()
	if err := crash.SetDir(base); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = crash.SetDir(t.TempDir()) })

	before := crash.Dir()
	err := crash.SetDir(dir)
	after := crash.Dir()
	if mustRejectDir(dir) && err == nil {
		t.Fatalf("accepted unsafe %q as %q", dir, after)
	}
	if err != nil {
		if after != before {
			t.Fatalf("rejected %q changed dir to %q", dir, after)
		}
		return
	}
	if !filepath.IsAbs(after) {
		t.Fatalf("accepted %q stored %q", dir, after)
	}
	for _, part := range strings.FieldsFunc(after, isPathSep) {
		if part == ".." {
			t.Fatalf("accepted %q has .. in %q", dir, after)
		}
	}
}

func isPathSep(r rune) bool {
	return r == '/' || r == '\\'
}
