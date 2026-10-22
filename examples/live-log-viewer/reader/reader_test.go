package reader

import (
	"context"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/chinmay-sawant/ownframe/examples/live-log-viewer/entry"
)

func testFile(t *testing.T, path string, pos int64) *File {
	t.Helper()

	pol := entry.DefaultPolicy()
	pol.Poll = 2 * time.Millisecond

	f, err := NewFile(FileOptions{Path: path, Position: pos, Policy: pol})
	if err != nil {
		t.Fatal(err)
	}

	t.Cleanup(func() { _ = f.Close() })

	return f
}

func writeLog(t *testing.T, path, s string) {
	t.Helper()

	if err := os.WriteFile(path, []byte(s), 0o600); err != nil {
		t.Fatal(err)
	}
}

func appendLog(t *testing.T, path string, b []byte) {
	t.Helper()

	f, err := os.OpenFile(path, os.O_APPEND|os.O_WRONLY, 0o600) //nolint:gosec // reason: test fixture under TempDir, not user input
	if err != nil {
		t.Fatal(err)
	}

	if _, err := f.Write(b); err != nil {
		t.Fatal(err)
	}

	_ = f.Close()
}

func drain(t *testing.T, r Reader, n int) []entry.RawRecord {
	t.Helper()

	var out []entry.RawRecord

	deadline := time.Now().Add(3 * time.Second)

	for len(out) < n && time.Now().Before(deadline) {
		ctx, cancel := context.WithTimeout(context.Background(), time.Second)

		b, err := r.Read(ctx)

		cancel()

		if err != nil {
			t.Fatal(err)
		}

		out = append(out, b.Records...)
	}

	if len(out) < n {
		t.Fatalf("got %d records, want %d", len(out), n)
	}

	return out
}

func logPath(t *testing.T) string {
	t.Helper()

	return filepath.Join(t.TempDir(), "app.log")
}
