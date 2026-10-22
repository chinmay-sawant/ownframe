package transfer

import (
	"os"
	"path/filepath"
	"time"
)

// FakeOptions configures the dummy transport.
type FakeOptions struct {
	// Dir is the directory the dummy transport writes into. Empty means
	// DummyDir().
	Dir string
	// Pace is the pause between chunks. Zero means 25 ms; tests set a
	// small value.
	Pace time.Duration
	// Chunk is the number of bytes per progress tick. Zero means 32 KiB.
	Chunk int64
}

// DummyDir is the dedicated temporary directory dummy jobs write to.
func DummyDir() string {
	return filepath.Join(os.TempDir(), "ownframe-download-manager-dummy")
}

// Fake is a deterministic transport for dummy mode. It never opens a
// network connection. The body size, progress steps, and one-time failures
// derive from the job ID, so a rerun shows the same behavior.
type Fake struct {
	pace  time.Duration
	chunk int64
}

// NewFake builds the dummy transport.
func NewFake(opts FakeOptions) *Fake {
	if opts.Pace <= 0 {
		opts.Pace = fakePace
	}

	if opts.Chunk <= 0 {
		opts.Chunk = fakeChunk
	}

	return &Fake{pace: opts.Pace, chunk: opts.Chunk}
}
