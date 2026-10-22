package ownframe

import (
	"context"
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// TestRunNilPageWritesReport checks the global fallback: a startup
// failure is stored in a report file, and the error keeps ErrNilPage
// without naming the file. LastReportPath exposes the file.
func TestRunNilPageWritesReport(t *testing.T) {
	dir := t.TempDir()
	if err := SetCrashDir(dir); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = SetCrashDir(t.TempDir()) })

	err := RunWithOptions(context.Background(), nil, WindowOptions{})
	if err == nil {
		t.Fatal("Run with nil page returned nil")
	}
	if !errors.Is(err, ErrNilPage) {
		t.Fatalf("err = %v, want ErrNilPage", err)
	}

	files, rerr := filepath.Glob(filepath.Join(dir, "*.txt"))
	if rerr != nil || len(files) != 1 {
		t.Fatalf("reports = %v, err = %v", files, rerr)
	}

	body, rerr := os.ReadFile(files[0])
	if rerr != nil {
		t.Fatal(rerr)
	}
	if !strings.Contains(string(body), "ownframe: nil page") {
		t.Fatalf("report body = %q", body)
	}
	if strings.Contains(err.Error(), files[0]) {
		t.Fatalf("error %q must not name the report file", err)
	}
	if got := LastReportPath(); got != files[0] {
		t.Fatalf("LastReportPath = %q, want %q", got, files[0])
	}
}
