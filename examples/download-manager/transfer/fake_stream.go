package transfer

import (
	"context"
	"fmt"
	"os"
	"time"
)

// stream writes the body from offset to size, reporting per chunk, and
// fails once mid-body for the jobs fakeFails marks.
func (f *Fake) stream(ctx context.Context, req Request, offset, size, total int64, report Reporter) error {
	stop := int64(-1)

	if fakeFails(req.JobID) && req.Attempt == 0 {
		stop = size * failNumer / failDenom
		if stop <= offset {
			stop = -1
		}
	}

	file, err := os.OpenFile(req.Partial, os.O_CREATE|os.O_WRONLY, transferFilePerm)
	if err != nil {
		return err
	}
	defer func() { _ = file.Close() }()

	if _, err := file.Seek(offset, 0); err != nil {
		return err
	}

	done := offset
	buf := make([]byte, f.chunk)

	for done < size {
		if stop >= 0 && done >= stop {
			return fmt.Errorf("transfer: fake connection reset at %d bytes", done)
		}

		if err := sleep(ctx, f.pace); err != nil {
			return err
		}

		n := int64(len(buf))
		if n > size-done {
			n = size - done
		}

		copy(buf, bodyBytes(done, n))

		if _, err := file.Write(buf[:n]); err != nil {
			return err
		}

		done += n

		if report != nil {
			report(Progress{Done: done, Total: total})
		}
	}

	return nil
}

// partialOffset returns the resume offset and whether the transfer resumes.
// A partial larger than the body restarts from zero.
func partialOffset(partial string, size int64) (int64, bool, error) {
	info, err := os.Stat(partial)
	if err != nil {
		if os.IsNotExist(err) {
			return 0, false, nil
		}

		return 0, false, err
	}

	offset := info.Size()
	if offset > size {
		if err := os.Truncate(partial, 0); err != nil {
			return 0, false, err
		}

		return 0, false, nil
	}

	return offset, offset > 0, nil
}

// sleep waits for d or until ctx ends.
func sleep(ctx context.Context, d time.Duration) error {
	if d <= 0 {
		return ctx.Err()
	}

	timer := time.NewTimer(d)
	defer timer.Stop()

	select {
	case <-ctx.Done():
		return ctx.Err()
	case <-timer.C:
		return nil
	}
}
