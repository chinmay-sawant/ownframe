package store

import (
	"os"
	"path/filepath"
	"testing"
)

// TestUnwritableDirectoryFailsWithoutReset reports the filesystem error.
func TestUnwritableDirectoryFailsWithoutReset(t *testing.T) {
	if os.Geteuid() == 0 {
		t.Skip("root ignores directory permissions")
	}

	parent := t.TempDir()
	if err := os.Chmod(parent, 0o500); err != nil { //nolint:gosec // reason: test makes TempDir read-only to prove Open fails.
		t.Fatal(err)
	}

	t.Cleanup(func() { _ = os.Chmod(parent, 0o700) }) //nolint:gosec // reason: test restores TempDir after read-only check.

	if _, err := Open(filepath.Join(parent, "sub")); err == nil {
		t.Fatal("read-only directory accepted")
	}
}
