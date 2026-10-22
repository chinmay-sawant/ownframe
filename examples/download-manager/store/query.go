package store

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"github.com/chinmay-sawant/ownframe/examples/download-manager/domain"
)

// Job reads one row.
func (s *Store) Job(ctx context.Context, id string) (domain.Job, error) {
	var job domain.Job

	err := s.do(ctx, func(ctx context.Context) error {
		row := s.db.QueryRowContext(ctx,
			`SELECT `+jobColumns+` FROM jobs WHERE id = ?`, id)

		var err error
		job, err = scanJob(row)

		return err
	})

	if errors.Is(err, sql.ErrNoRows) {
		return domain.Job{}, fmt.Errorf("%w: %s", ErrNotFound, id)
	}

	return job, err
}

// ActiveJobs returns queued, running, and paused jobs, oldest first.
func (s *Store) ActiveJobs(ctx context.Context) ([]domain.Job, error) {
	return s.queryJobs(ctx,
		`SELECT `+jobColumns+` FROM jobs
		 WHERE state IN ('queued','running','paused')
		 ORDER BY created_ms, id`)
}

// queryJobs runs one job SELECT on the worker and collects the rows. Rows
// close before the worker returns, so no second query can deadlock.
func (s *Store) queryJobs(ctx context.Context, query string, args ...any) ([]domain.Job, error) {
	var out []domain.Job

	err := s.do(ctx, func(ctx context.Context) error {
		rows, err := s.db.QueryContext(ctx, query, args...)
		if err != nil {
			return err
		}
		defer func() { _ = rows.Close() }()

		for rows.Next() {
			job, err := scanJob(rows)
			if err != nil {
				return err
			}

			out = append(out, job)
		}

		return rows.Err()
	})

	return out, err
}
