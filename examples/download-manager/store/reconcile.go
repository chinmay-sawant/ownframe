package store

import (
	"context"

	"github.com/chinmay-sawant/ownframe/examples/download-manager/domain"
)

// Reconcile repairs the gap between rows and files after a crash. Running
// rows become paused or completed; completed rows with a missing file
// become failed; a partial longer than expected restarts. It must run once
// before scheduler.Recover.
func (s *Store) Reconcile(ctx context.Context) (Report, error) {
	var rep Report

	err := s.do(ctx, func(ctx context.Context) error {
		jobs, err := s.reconcileSet(ctx)
		if err != nil {
			return err
		}

		for _, job := range jobs {
			if err := s.reconcileOne(ctx, job, &rep); err != nil {
				return err
			}
		}

		return nil
	})

	return rep, err
}

// reconcileSet reads every row that can need repair: active states plus
// completed rows, whose final file may have vanished.
func (s *Store) reconcileSet(ctx context.Context) ([]domain.Job, error) {
	rows, err := s.db.QueryContext(ctx,
		`SELECT `+jobColumns+` FROM jobs
		 WHERE state IN ('running','paused','queued','completed')
		 ORDER BY created_ms, id`)
	if err != nil {
		return nil, err
	}
	defer func() { _ = rows.Close() }()

	var out []domain.Job

	for rows.Next() {
		job, err := scanJob(rows)
		if err != nil {
			return nil, err
		}

		out = append(out, job)
	}

	return out, rows.Err()
}

// reconcileOne applies the rule for one row.
func (s *Store) reconcileOne(ctx context.Context, job domain.Job, rep *Report) error {
	switch job.State {
	case domain.StateRunning:
		return s.reconcileRunning(ctx, job, rep)
	case domain.StatePaused, domain.StateQueued:
		return s.reconcilePartial(ctx, job, rep)
	case domain.StateCompleted:
		return s.reconcileFinished(ctx, job, rep)
	}

	return nil
}
