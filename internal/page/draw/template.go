package draw

import (
	"bytes"
	"io"
	"sync"
	"time"
)

// bufferPool reuses template execution scratch buffers across frames so a
// redraw does not Grow a fresh backing array every frame. Unlike
// strings.Builder.Reset (which drops its backing slice to preserve string
// immutability), bytes.Buffer.Reset retains its backing array while
// Execute / ExecuteReuse / ExecuteWriter hand back a fresh caller-owned
// string or byte copy (or reuse an already-immutable prev string when the
// output is byte-identical).
//
// Retention is capped (see maxPooledScratchBytes): oversized buffers are
// dropped instead of pooled, so one large template cannot pin its high-water
// backing array for the life of the process.
//
//nolint:gochecknoglobals // reason: intentional process scratch pool, Reset on release.
var bufferPool sync.Pool

// maxPooledScratchBytes caps how much backing memory a released scratch
// buffer may retain. Rationale: the pool is process-wide, so without a
// cap one large execution (template bodies reach ~600KB in the golden and
// bench fixtures) grows the pooled backing array to that high-water mark
// and sync.Pool keeps it alive for the rest of the process, multiplied by
// GOMAXPROCS shards. 64 KiB covers the common small-frame reuse case with
// one pooled allocation while letting rare large frames allocate fresh and
// be collected, bounding retained heap to ~cap * GOMAXPROCS. Scratch only:
// the cap never changes rendered output, only whether the next frame
// reuses or reallocates.
const maxPooledScratchBytes = 64 * 1024

func acquireBuffer(hint int) *bytes.Buffer {
	if v, ok := bufferPool.Get().(*bytes.Buffer); ok && v != nil {
		v.Reset()
		if hint > 0 {
			v.Grow(hint)
		}

		return v
	}

	var buf bytes.Buffer
	if hint > 0 {
		buf.Grow(hint)
	}

	return &buf
}

func releaseBuffer(buf *bytes.Buffer) {
	if buf == nil {
		return
	}

	buf.Reset()
	if buf.Cap() > maxPooledScratchBytes {
		return
	}
	bufferPool.Put(buf)
}

// Execute runs run into a pooled bytes.Buffer grown to hint bytes when
// hint > 0 and returns a caller-owned string copy. When track is true it
// times the run and stores the duration in last. A nil last skips the store.
func Execute(track bool, hint int, last *time.Duration, run func(*bytes.Buffer) error) (string, error) {
	return ExecuteReuse(track, hint, "", last, run)
}

// ExecuteReuse runs run into a pooled bytes.Buffer grown to hint bytes when
// hint > 0. When prev is non-empty and matches the rendered bytes, it returns
// prev without allocating a new string. Otherwise it returns a caller-owned
// string copy so the pooled buffer can be Reset and reused without aliasing.
func ExecuteReuse(
	track bool,
	hint int,
	prev string,
	last *time.Duration,
	run func(*bytes.Buffer) error,
) (string, error) {
	start := Clock(track)

	buf := acquireBuffer(hint)

	err := run(buf)
	if last != nil {
		Note(track, last, start)
	}
	if err != nil {
		releaseBuffer(buf)

		return "", err
	}

	raw := buf.Bytes()
	if prev != "" && len(raw) == len(prev) && string(raw) == prev {
		releaseBuffer(buf)

		return prev, nil
	}

	out := buf.String()
	releaseBuffer(buf)

	return out, nil
}

// ExecuteWriter runs run into a pooled buffer grown to hint bytes when
// hint > 0 and returns the bytes as a caller-owned copy (copy-on-write:
// the pooled buffer is released for the next frame, the returned slice
// is a fresh allocation the caller may retain). A changed frame pays one
// copy instead of Grow plus String. Timing follows Execute.
func ExecuteWriter(track bool, hint int, last *time.Duration, run func(io.Writer) error) ([]byte, error) {
	start := Clock(track)

	buf := acquireBuffer(hint)

	err := run(buf)
	if last != nil {
		Note(track, last, start)
	}
	if err != nil {
		releaseBuffer(buf)

		return nil, err
	}

	out := bytes.Clone(buf.Bytes())
	releaseBuffer(buf)

	return out, nil
}

// HashString hashes s with FNV-1a without converting it to a byte slice.
// Keys and the page loop use it to compare sources by u64 first, so an
// unchanged frame skips the full-source memcmp.
func HashString(s string) uint64 {
	const (
		offset64 = 14695981039346656037
		prime64  = 1099511628211
	)

	hash := uint64(offset64)
	for i := 0; i < len(s); i++ {
		hash ^= uint64(s[i])
		hash *= prime64
	}

	return hash
}
