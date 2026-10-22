package transfer

import (
	"context"
	"errors"
	"os"
	"path/filepath"
	"testing"

	"github.com/chinmay-sawant/ownframe/examples/download-manager/fixture"
)

// TestHTTPPartialUnwritable reports a write failure instead of losing the
// bytes. A file blocks the partial's parent path.
func TestHTTPPartialUnwritable(t *testing.T) {
	srv := fixture.NewServer()
	defer srv.Close()

	dir := t.TempDir()
	blocker := filepath.Join(dir, "blocked")

	if err := os.WriteFile(blocker, []byte("x"), 0o600); err != nil {
		t.Fatal(err)
	}

	dest := filepath.Join(dir, "out.bin")
	partial := filepath.Join(blocker, "out.bin"+PartialSuffix)

	_, err := newHTTP(t).Download(context.Background(), Request{
		URL: srv.URL + fixture.PathOK, Dest: dest, Partial: partial,
	}, nil)
	if !errors.Is(err, ErrPartialUnwritable) {
		t.Fatalf("want ErrPartialUnwritable, got %v", err)
	}

	if _, statErr := os.Stat(dest); !os.IsNotExist(statErr) {
		t.Error("failed transfer finalized a destination")
	}
}
