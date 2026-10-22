package store

import (
	"context"
	"errors"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/chinmay-sawant/ownframe/examples/live-log-viewer/entry"
)

func fileSource(t *testing.T, st *Store, dir, content string) entry.Source {
	t.Helper()

	sess, err := st.EnsureSession(bg(), "files", "file")
	if err != nil {
		t.Fatal(err)
	}

	logPath := filepath.Join(dir, "app.log")
	if err := os.WriteFile(logPath, []byte(content), 0o600); err != nil {
		t.Fatal(err)
	}

	src, err := st.AddSource(bg(), SourceSpec{
		Session: sess.ID, Path: logPath, Label: "app",
	})
	if err != nil {
		t.Fatal(err)
	}

	return src
}

func startFileIngest(t *testing.T, st *Store, src entry.Source) (*Ingestor, context.CancelFunc, chan error) {
	t.Helper()

	ctx, cancel := context.WithCancel(bg())
	done := make(chan error, 1)

	in, err := st.Ingest(ctx, src.ID, testPolicy())
	if err != nil {
		t.Fatal(err)
	}

	go func() { done <- in.Run(ctx) }()

	return in, cancel, done
}

func TestIngestFileShutdownPartial(t *testing.T) {
	dir := t.TempDir()

	st, err := Open(dir)
	if err != nil {
		t.Fatal(err)
	}

	defer func() { _ = st.Close() }()

	src := fileSource(t, st, dir, "INFO first\nINFO partial")

	in, cancel, done := startFileIngest(t, st, src)

	if got := waitForEntries(t, st, 1, 3*time.Second); got < 1 {
		cancel()
		if err := <-done; err != nil && !errors.Is(err, context.Canceled) {
			t.Fatal(err)
		}
		t.Fatalf("entries = %d", got)
	}

	cancel()
	if err := <-done; err != nil && !errors.Is(err, context.Canceled) {
		t.Fatal(err)
	}
	if err := in.Close(); err != nil {
		t.Fatal(err)
	}

	pos, err := st.LastPosition(bg(), src.ID)
	if err != nil {
		t.Fatal(err)
	}

	if !pos.OK || !pos.Partial {
		t.Fatalf("shutdown did not store the partial line: %+v", pos)
	}
}
