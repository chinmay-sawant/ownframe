package ownframe

import (
	"context"
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
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

// TestRunWithCancelledContextDoesNotWriteReport checks that stopping Run
// with a canceled context or a deadline returns that error and leaves the
// crash directory empty. LastReportPath is process-global, so the test
// records it and requires the call to leave it unchanged.
func TestRunWithCancelledContextDoesNotWriteReport(t *testing.T) {
	dir := t.TempDir()
	if err := SetCrashDir(dir); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = SetCrashDir(t.TempDir()) })

	page, err := New(Config{HTML: "<p>hi</p>", Width: 200, Height: 100})
	if err != nil {
		t.Fatal(err)
	}

	cases := []struct {
		name string
		ctx  context.Context
		want error
	}{
		{name: "canceled", ctx: canceledContext(), want: context.Canceled},
		{name: "deadline", ctx: deadlineContext(), want: context.DeadlineExceeded},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			before := LastReportPath()
			err := RunWithOptions(tc.ctx, page, WindowOptions{})
			if !errors.Is(err, tc.want) {
				t.Fatalf("err = %v, want %v", err, tc.want)
			}

			files, rerr := filepath.Glob(filepath.Join(dir, "*.txt"))
			if rerr != nil || len(files) != 0 {
				t.Fatalf("reports = %v, err = %v", files, rerr)
			}
			if got := LastReportPath(); got != before {
				t.Fatalf("LastReportPath = %q, want %q", got, before)
			}
		})
	}
}

func canceledContext() context.Context {
	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	return ctx
}

func deadlineContext() context.Context {
	ctx, cancel := context.WithDeadline(context.Background(), time.Unix(0, 0))
	cancel()

	return ctx
}
