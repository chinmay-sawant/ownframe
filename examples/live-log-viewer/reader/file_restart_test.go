package reader

import (
	"os"
	"testing"
	"time"

	"github.com/chinmay-sawant/ownframe/examples/live-log-viewer/entry"
)

func storedFile(t *testing.T, path string, gen int64) *File {
	t.Helper()

	fi, err := os.Stat(path)
	if err != nil {
		t.Fatal(err)
	}

	id, err := pathIdentity(path)
	if err != nil {
		t.Fatal(err)
	}

	f, err := os.Open(path) //nolint:gosec // reason: test reads a TempDir log file, not user input
	if err != nil {
		t.Fatal(err)
	}

	head, headLen := fileHeadHash(f, fi.Size(), 0)

	_ = f.Close()

	pol := entry.DefaultPolicy()
	pol.Poll = 2 * time.Millisecond

	file, err := NewFile(FileOptions{
		Path: path, Position: fi.Size(), Generation: gen,
		Identity: id, HeadHash: head, HeadLen: headLen, Policy: pol,
	})
	if err != nil {
		t.Fatal(err)
	}

	t.Cleanup(func() { _ = file.Close() })

	return file
}

func TestFileRestartAfterRotation(t *testing.T) {
	path := logPath(t)
	writeLog(t, path, "old\n")

	r := storedFile(t, path, 1)

	if err := os.Remove(path); err != nil {
		t.Fatal(err)
	}

	writeLog(t, path, "new\n")

	recs := drain(t, r, 1)
	if recs[0].Generation != 2 || recs[0].Offset != 0 || string(recs[0].Data) != "new" {
		t.Fatalf("record = %+v", recs[0])
	}
}

func TestFileRestartAfterTruncation(t *testing.T) {
	path := logPath(t)
	writeLog(t, path, "first line\nsecond line\n")

	r := storedFile(t, path, 1)

	writeLog(t, path, "tiny\n")

	recs := drain(t, r, 1)
	if recs[0].Generation != 2 || recs[0].Offset != 0 || string(recs[0].Data) != "tiny" {
		t.Fatalf("record = %+v", recs[0])
	}
}
