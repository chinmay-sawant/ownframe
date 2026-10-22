package reader

import (
	"context"
	"io"
	"time"

	"github.com/chinmay-sawant/ownframe/examples/live-log-viewer/entry"
)

// open opens the path with the platform sharing flags and records the file
// identity from the handle, not the path, so a later rename is visible.
func (r *File) open() error {
	f, err := openShared(r.opts.Path)
	if err != nil {
		return err
	}

	if fi, err := f.Stat(); err == nil {
		r.size = fi.Size()

		if id, err := handleIdentity(f); err == nil && id != "" {
			r.identity = id
		}

		head, headLen := fileHeadHash(f, fi.Size(), r.opts.HeadLen)
		if r.opts.HeadHash != 0 && head != 0 && head != r.opts.HeadHash {
			// Same path and possibly a recycled inode, different content:
			// start a new generation from the first byte.
			r.gen++
			r.pos = 0
			r.reset()
			r.reopened = true
		}

		if head != 0 {
			r.opts.HeadHash, r.opts.HeadLen = head, headLen
		}

		r.head, r.headLen = head, headLen
	}

	r.f = f
	r.state = entry.StateLive

	return nil
}

// readChunk reads at most max bytes at the current position using ReadAt, so
// a shorter file never blocks or seeks past rewritten bytes.
func (r *File) readChunk(max int) ([]byte, error) {
	if r.f == nil {
		return nil, nil
	}

	if max <= 0 || max > r.pol.BatchBytes {
		max = r.pol.BatchBytes
	}

	if max > readMaxCap {
		max = readMaxCap
	}

	if cap(r.chunk) < max {
		r.chunk = make([]byte, max)
	}

	buf := r.chunk[:max]

	n, err := r.f.ReadAt(buf, r.pos)
	if n > 0 {
		r.pos += int64(n)
	}

	if err != nil && err != io.EOF {
		return nil, err
	}

	return buf[:n], nil
}

func (r *File) wait(ctx context.Context) error {
	d := r.pol.Poll
	if d <= 0 {
		return ctx.Err()
	}

	if r.opts.Sleep != nil {
		return r.opts.Sleep(ctx, d)
	}

	t := time.NewTimer(d)
	defer t.Stop()

	select {
	case <-ctx.Done():
		return ctx.Err()
	case <-t.C:
		return nil
	}
}
