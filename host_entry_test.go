package ownframe

import (
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// withCrashDir points crash reports at a temp directory for one test.
func withCrashDir(t *testing.T) string {
	t.Helper()
	dir := t.TempDir()
	if err := SetCrashDir(dir); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = SetCrashDir(t.TempDir()) })

	return dir
}

func reportFiles(t *testing.T, dir string) []string {
	t.Helper()
	files, err := filepath.Glob(filepath.Join(dir, "*.txt"))
	if err != nil {
		t.Fatal(err)
	}

	return files
}

// TestBindMobileNilPage checks a nil page returns ErrNilPage and does not
// panic. PrepareMobile rejects nil, so BindMobile needs no extra check.
// A recovered panic would be ErrPanicked and a crash file.
func TestBindMobileNilPage(t *testing.T) {
	dir := withCrashDir(t)

	err := BindMobile(t.Context(), nil)
	if err == nil {
		t.Fatal("BindMobile(nil) returned nil")
	}
	if !errors.Is(err, ErrNilPage) {
		t.Fatalf("err = %v, want ErrNilPage", err)
	}
	if errors.Is(err, ErrPanicked) {
		t.Fatalf("nil page panicked: %v", err)
	}
	if files := reportFiles(t, dir); len(files) != 0 {
		t.Fatalf("nil page wrote reports %v", files)
	}
}

// runPanicHTML is one filled box, so prepare commits a display and runs
// the after-redraw callback. The callback panics there, still inside
// prepare, and Run returns before it opens a window. A template method
// panic does not reach savePanic: html/template turns it into an error.
// A handler panic runs only inside the Ebiten loop.
const runPanicHTML = `<style>div { width: 8px; height: 8px; background: #ff0000; }</style><div id="a"></div>`

// TestRunPagePanicWritesReport checks Run's savePanic on a page panic.
// TestPanicReportKeepsChain checks the error chain and not the crash file.
func TestRunPagePanicWritesReport(t *testing.T) {
	dir := withCrashDir(t)
	page, err := New(Config{
		HTML:   runPanicHTML,
		Width:  1,
		Height: 1,
	})
	if err != nil {
		t.Fatal(err)
	}
	page.SetAfterRedraw(func() { panic("host page panic") })

	err = Run(t.Context(), page)
	if err == nil {
		t.Fatal("Run returned nil")
	}
	if !errors.Is(err, ErrPanicked) {
		t.Fatalf("err = %v, want ErrPanicked", err)
	}
	if strings.Contains(err.Error(), dir) {
		t.Fatalf("error %q names the crash directory", err)
	}

	files := reportFiles(t, dir)
	if len(files) != 1 {
		t.Fatalf("reports = %v", files)
	}
	if got := LastReportPath(); got != files[0] {
		t.Fatalf("LastReportPath = %q, want %q", got, files[0])
	}
	body, rerr := os.ReadFile(files[0])
	if rerr != nil {
		t.Fatal(rerr)
	}
	if !strings.Contains(string(body), "host page panic") {
		t.Fatalf("report body = %q", body)
	}
}
