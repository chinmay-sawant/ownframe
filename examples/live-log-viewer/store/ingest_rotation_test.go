package store

import (
	"context"
	"errors"
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestIngestRestartAfterRotation(t *testing.T) {
	dir := t.TempDir()

	st, err := Open(dir)
	if err != nil {
		t.Fatal(err)
	}

	src := fileSource(t, st, dir, "old entry\n")

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

	before, err := st.Source(bg(), src.ID)
	if err != nil {
		t.Fatal(err)
	}

	if before.HeadHash == 0 {
		t.Fatal("head hash was not committed")
	}

	logPath := filepath.Join(dir, "app.log")

	if err := os.Remove(logPath); err != nil {
		t.Fatal(err)
	}

	if err := os.WriteFile(logPath, []byte("new entry\n"), 0o600); err != nil {
		t.Fatal(err)
	}

	in2, cancel2, done2 := startFileIngest(t, st, src)

	if got := waitForEntries(t, st, 2, 3*time.Second); got < 2 {
		cancel2()
		if err := <-done2; err != nil && !errors.Is(err, context.Canceled) {
			t.Fatal(err)
		}
		t.Fatalf("entries after rotation = %d", got)
	}

	cancel2()
	if err := <-done2; err != nil && !errors.Is(err, context.Canceled) {
		t.Fatal(err)
	}
	if err := in2.Close(); err != nil {
		t.Fatal(err)
	}

	after, err := st.Source(bg(), src.ID)
	if err != nil {
		t.Fatal(err)
	}

	if after.Generation != before.Generation+1 {
		t.Fatalf("generation %d -> %d", before.Generation, after.Generation)
	}

	entries := allEntries(t, st, Query{Source: &src.ID})
	if len(entries) != 2 {
		t.Fatalf("entries = %d, want 2", len(entries))
	}

	if entries[1].Message != "new entry" {
		t.Fatalf("second entry = %q", entries[1].Message)
	}

	if entries[1].Generation != after.Generation {
		t.Fatalf("entry generation = %d", entries[1].Generation)
	}
}
