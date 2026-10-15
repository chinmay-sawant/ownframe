package crash_test

import (
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"

	"github.com/chinmay-sawant/ownframe/internal/crash"
)

func TestWriteStoresTitleAndReason(t *testing.T) {
	dir := filepath.Join(t.TempDir(), "reports")
	if err := crash.SetDir(dir); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = crash.SetDir(t.TempDir()) })

	path, err := crash.Write("Sign in", "nil pointer")
	if err != nil {
		t.Fatal(err)
	}

	body := readReport(t, path)
	if !strings.Contains(body, "Sign in") || !strings.Contains(body, "nil pointer") {
		t.Fatalf("body = %s", body)
	}

	dirInfo, err := os.Stat(dir)
	if err != nil {
		t.Fatal(err)
	}
	if dirInfo.Mode().Perm() != 0700 {
		t.Fatalf("dir mode = %v, want 0700", dirInfo.Mode().Perm())
	}
	fileInfo, err := os.Stat(path)
	if err != nil {
		t.Fatal(err)
	}
	if fileInfo.Mode().Perm() != 0600 {
		t.Fatalf("file mode = %v, want 0600", fileInfo.Mode().Perm())
	}

	if !strings.Contains(body, runtime.Version()) || !strings.Contains(body, "goroutine ") {
		t.Fatalf("body = %s", body)
	}

	second, err := crash.Write("Sign in", "nil pointer")
	if err != nil {
		t.Fatal(err)
	}

	if second == path {
		t.Fatal("second write reused the path")
	}

	_ = readReport(t, second)
}

func readReport(t *testing.T, path string) string {
	t.Helper()

	info, err := os.Stat(path)
	if err != nil {
		t.Fatal(err)
	}

	if info.IsDir() {
		t.Fatalf("%s is a directory", path)
	}

	base := filepath.Base(path)
	if len(base) < len("20060102-150405") || base[8] != '-' {
		t.Fatalf("name = %s", base)
	}

	raw, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}

	return string(raw)
}
