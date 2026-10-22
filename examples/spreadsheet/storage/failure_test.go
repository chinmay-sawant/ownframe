package storage

import (
	"context"
	"os"
	"path/filepath"
	"testing"
)

func TestReadOnlyDirFailsCleanly(t *testing.T) {
	if os.Geteuid() == 0 {
		t.Skip("root ignores directory permissions")
	}

	dir := filepath.Join(t.TempDir(), "readonly")

	if err := os.MkdirAll(dir, 0o750); err != nil {
		t.Fatal(err)
	}

	if err := os.Chmod(dir, 0o500); err != nil { //nolint:gosec // reason: test makes TempDir read-only to verify failure, then restores it
		t.Fatal(err)
	}

	t.Cleanup(func() { _ = os.Chmod(dir, 0o750) }) //nolint:gosec // reason: test restores TempDir permissions for cleanup

	if _, err := Open(dir); err == nil {
		t.Fatal("Open in a read-only directory succeeded")
	}
}

func TestReadOnlyFileFailsOnWrite(t *testing.T) {
	if os.Geteuid() == 0 {
		t.Skip("root ignores file permissions")
	}

	st, dir := openDir(t)

	if _, err := st.CreateWorkbook(context.Background(), "book"); err != nil {
		t.Fatal(err)
	}

	_ = st.Close()

	path := filepath.Join(dir, File)

	if err := os.Chmod(path, 0o400); err != nil {
		t.Fatal(err)
	}

	t.Cleanup(func() { _ = os.Chmod(path, 0o600) })

	st, err := Open(dir)
	if err != nil {
		return // surfaced at open, which is also fine
	}

	defer func() { _ = st.Close() }()

	if _, err := st.CreateWorkbook(context.Background(), "write"); err == nil {
		t.Fatal("write on a read-only database succeeded")
	}
}
