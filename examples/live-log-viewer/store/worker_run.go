package store

import (
	"context"
	"database/sql"
	"time"
)

// longJobTimeout is the deadline for maintenance work such as seeding,
// pruning, exporting, and backing up. Point queries keep the shorter
// configured timeout.
const longJobTimeout = 60 * time.Second

// runJob sends fn to the worker with the configured query deadline.
func runJob[T any](s *Store, ctx context.Context, fn func(context.Context, *sql.DB) (T, error)) (T, error) {
	return runJobFor(s, ctx, s.timeout, fn)
}

// runJobLong sends fn to the worker with the maintenance deadline.
func runJobLong[T any](s *Store, ctx context.Context, fn func(context.Context, *sql.DB) (T, error)) (T, error) {
	d := s.timeout
	if d < longJobTimeout {
		d = longJobTimeout
	}

	return runJobFor(s, ctx, d, fn)
}

// runJobFor waits for the worker, a query deadline, or a closed store,
// whichever comes first.
func runJobFor[T any](s *Store, ctx context.Context, d time.Duration, fn func(context.Context, *sql.DB) (T, error)) (T, error) {
	var zero T

	if s == nil || s.closed.Load() {
		return zero, ErrClosed
	}

	if d <= 0 {
		d = storeDefTimeout
	}

	ctx, cancel := context.WithTimeout(ctx, d)
	defer cancel()

	j := &job{
		ctx:  ctx,
		resp: make(chan jobResult, 1),
		fn: func(c context.Context, db *sql.DB) (any, error) {
			return fn(c, db)
		},
	}

	select {
	case s.jobs <- j:
	case <-ctx.Done():
		return zero, ctx.Err()
	case <-s.stop:
		return zero, ErrClosed
	}

	select {
	case r := <-j.resp:
		if r.err != nil {
			return zero, r.err
		}

		v, ok := r.val.(T)
		if !ok {
			return zero, nil
		}

		return v, nil
	case <-ctx.Done():
		return zero, ctx.Err()
	case <-s.stop:
		return zero, ErrClosed
	}
}
