package draw

import (
	"bytes"
	"io"
	"strings"
	"sync"
	"time"
)

// builderPool reuses template builders across frames so a changed frame
// does not Grow a fresh buffer every redraw. Buffers are Reset before
// they go back, so a pooled builder never carries the previous frame.
// A process-wide pool is intended: templates execute on the page loop
// and the pool only recycles scratch memory, never page state.
//
// Retention is capped (see maxPooledScratchBytes): oversized builders
// are dropped instead of pooled, so one large template cannot pin its
// high-water backing array for the life of the process.
//
//nolint:gochecknoglobals // reason: intentional process scratch pool, Reset on release.
var builderPool sync.Pool

// bufferPool is the bytes.Buffer twin of builderPool, backing ExecuteWriter.
// It lets a template execution return bytes with one copy: the pooled
// buffer is cloned once for the caller, then reused. Oversized buffers
// are dropped on release under the same cap as builderPool.
//
//nolint:gochecknoglobals // reason: intentional process scratch pool, Reset on release.
var bufferPool sync.Pool

// maxPooledScratchBytes caps how much backing memory a released scratch
// buffer may retain. Rationale: the pools are process-wide, so without a
// cap one large execution (template bodies reach ~600KB in the golden and
// bench fixtures) grows the pooled backing array to that high-water mark
// and sync.Pool keeps it alive for the rest of the process, multiplied by
// GOMAXPROCS shards. 64 KiB covers the common small-frame reuse case with
// one pooled allocation while letting rare large frames allocate fresh and
// be collected, bounding retained heap to ~cap * GOMAXPROCS. Scratch only:
// the cap never changes rendered output, only whether the next frame
// reuses or reallocates.
const maxPooledScratchBytes = 64 * 1024

func acquireBuilder(hint int) *strings.Builder {
	if v, ok := builderPool.Get().(*strings.Builder); ok && v != nil {
		v.Reset()
		if hint > 0 {
			v.Grow(hint)
		}

		return v
	}

	var body strings.Builder
	if hint > 0 {
		body.Grow(hint)
	}

	return &body
}

func releaseBuilder(b *strings.Builder) {
	if b == nil {
		return
	}

	b.Reset()
	if b.Cap() > maxPooledScratchBytes {
		return
	}
	builderPool.Put(b)
}

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

// Execute runs run into a pooled builder grown to hint bytes when hint > 0.
// When track is true it times the run and stores the duration in last.
// A nil last skips the store. run receives the builder.
func Execute(track bool, hint int, last *time.Duration, run func(*strings.Builder) error) (string, error) {
	start := Clock(track)

	body := acquireBuilder(hint)

	err := run(body)
	if last != nil {
		Note(track, last, start)
	}
	if err != nil {
		releaseBuilder(body)

		return "", err
	}

	out := body.String()
	releaseBuilder(body)

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
