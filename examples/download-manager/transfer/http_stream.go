package transfer

import (
	"context"
	"fmt"
	"io"
	"math"
	"net/http"
	"os"
	"sync/atomic"
	"time"
)

// receive writes the response body to the partial file and reports
// progress. The stall guard cancels the request context when no byte
// arrives within the idle window.
func (h *HTTP) receive(ctx context.Context, cancel context.CancelFunc, req Request, resp *http.Response, p plan, report Reporter) (Outcome, error) {
	file, err := openPartial(req.Partial, p)
	if err != nil {
		return Outcome{}, err
	}

	var stalled int32

	timer := time.AfterFunc(h.stall, func() {
		atomic.StoreInt32(&stalled, 1)
		cancel()
	})
	defer timer.Stop()

	reset := func() { timer.Reset(h.stall) }

	done, err := pump(ctx, resp.Body, file, p, report, &stalled, reset)
	closeErr := file.Close()

	if err != nil {
		return Outcome{}, err
	}

	if closeErr != nil {
		return Outcome{}, closeErr
	}

	if p.total >= 0 && done != p.total {
		return Outcome{}, fmt.Errorf("%w: read %d of %d bytes",
			ErrBadStatus, done, p.total)
	}

	return h.complete(req, p, Outcome{Bytes: done, Resumed: p.resumed})
}

// pump copies body into file, resetting the stall timer on every read.
func pump(ctx context.Context, body io.Reader, file *os.File, p plan, report Reporter, stalled *int32, reset func()) (int64, error) {
	buf := make([]byte, streamChunk)
	done := p.start

	for {
		n, readErr := body.Read(buf)
		if n > 0 {
			if int64(n) > math.MaxInt64-done {
				return done, fmt.Errorf("transfer: byte counter overflow")
			}

			if _, err := file.Write(buf[:n]); err != nil {
				return done, err
			}

			done += int64(n)
			reset()

			if report != nil {
				report(Progress{Done: done, Total: p.total})
			}
		}

		if readErr == io.EOF {
			return done, nil
		}

		if readErr != nil {
			if atomic.LoadInt32(stalled) == 1 {
				return done, fmt.Errorf("%w: %w", ErrStall, readErr)
			}

			if err := ctx.Err(); err != nil {
				return done, err
			}

			return done, readErr
		}
	}
}
