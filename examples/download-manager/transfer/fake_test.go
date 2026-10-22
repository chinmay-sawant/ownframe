package transfer

import (
	"bytes"
	"context"
	"os"
	"path/filepath"
	"testing"
	"time"
)

// fastFake keeps tests quick: one big chunk and no real pause.
func fastFake() *Fake {
	return NewFake(FakeOptions{Pace: time.Nanosecond, Chunk: 1 << 20})
}

// fakePaths returns a fresh dest and partial pair.
func fakePaths(t *testing.T) (string, string) {
	t.Helper()
	dir := t.TempDir()
	dest := filepath.Join(dir, "out.bin")

	return dest, PartialPath(dest)
}

// readAll reads a file or fails the test.
func readAll(t *testing.T, path string) []byte {
	t.Helper()

	data, err := os.ReadFile(path) //nolint:gosec // reason: test reads back the fixture file it wrote.
	if err != nil {
		t.Fatal(err)
	}

	return data
}

// TestFakeCompletes writes the deterministic body and finalizes it.
func TestFakeCompletes(t *testing.T) {
	dest, partial := fakePaths(t)
	id := "complete-1"

	out, err := fastFake().Download(context.Background(), Request{
		JobID: id, Dest: dest, Partial: partial,
	}, nil)
	if err != nil {
		t.Fatal(err)
	}

	size := FakeSize(id)
	if out.Bytes != size {
		t.Errorf("bytes %d, want %d", out.Bytes, size)
	}

	if out.Resumed {
		t.Error("fresh transfer reported resumed")
	}

	if got := readAll(t, dest); !bytes.Equal(got, bodyBytes(0, size)) {
		t.Error("body mismatch")
	}

	if _, err := os.Stat(partial); !os.IsNotExist(err) {
		t.Error("partial still present after finalize")
	}
}

// TestFakeResumesFromPartial appends from the existing offset.
func TestFakeResumesFromPartial(t *testing.T) {
	dest, partial := fakePaths(t)
	id := "resume-1"

	if err := os.WriteFile(partial, bodyBytes(0, 1000), 0o600); err != nil {
		t.Fatal(err)
	}

	out, err := fastFake().Download(context.Background(), Request{
		JobID: id, Dest: dest, Partial: partial,
	}, nil)
	if err != nil {
		t.Fatal(err)
	}

	if !out.Resumed {
		t.Error("resume not reported")
	}

	size := FakeSize(id)
	if got := readAll(t, dest); !bytes.Equal(got, bodyBytes(0, size)) {
		t.Error("resumed body mismatch")
	}
}
