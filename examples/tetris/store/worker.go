package store

import (
	"context"
	"database/sql"
)

// request is one job for the worker.
type request struct {
	ctx  context.Context
	fn   func(context.Context, *sql.DB) error
	done chan struct{}
	err  error
}

// do sends fn to the worker and waits for it or for the context. Callers
// never run SQL themselves. A full queue returns ErrBusy instead of
// blocking the caller.
func (s *Store) do(ctx context.Context, fn func(context.Context, *sql.DB) error) error {
	if ctx == nil {
		ctx = context.Background()
	}

	if _, ok := ctx.Deadline(); !ok {
		var cancel context.CancelFunc

		ctx, cancel = context.WithTimeout(ctx, QueryTimeout)
		defer cancel()
	}

	req := &request{ctx: ctx, fn: fn, done: make(chan struct{})}

	s.mu.Lock()
	if s.closed {
		s.mu.Unlock()

		return ErrClosed
	}

	select {
	case s.reqs <- req:
	default:
		s.mu.Unlock()

		return ErrBusy
	}
	s.mu.Unlock()

	select {
	case <-req.done:
		return req.err
	case <-ctx.Done():
		return ctx.Err()
	}
}

// worker runs requests one at a time until Close.
func (s *Store) worker() {
	defer close(s.done)

	for {
		select {
		case req := <-s.reqs:
			s.run(req)
		case <-s.quit:
			s.drain()
			_ = s.db.Close()
			return
		}
	}
}

// drain finishes requests already queued at Close.
func (s *Store) drain() {
	for {
		select {
		case req := <-s.reqs:
			s.run(req)
		default:
			return
		}
	}
}

// run executes one request unless its deadline already passed.
func (s *Store) run(req *request) {
	if err := req.ctx.Err(); err != nil {
		req.err = err
	} else {
		req.err = req.fn(req.ctx, s.db)
	}

	close(req.done)
}
