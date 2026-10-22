package store

import (
	"context"

	"github.com/chinmay-sawant/ownframe/examples/download-manager/domain"
)

// Summary returns the per-state counts. The UI calls it at a bounded rate,
// never once per progress event.
func (s *Store) Summary(ctx context.Context) (Counts, error) {
	counts := Counts{ByState: map[domain.State]int{}}

	err := s.do(ctx, func(ctx context.Context) error {
		rows, err := s.db.QueryContext(ctx,
			`SELECT state, count(*) FROM jobs GROUP BY state`)
		if err != nil {
			return err
		}
		defer func() { _ = rows.Close() }()

		for rows.Next() {
			var (
				state string
				n     int
			)

			if err := rows.Scan(&state, &n); err != nil {
				return err
			}

			counts.ByState[domain.State(state)] = n
			counts.Total += n
		}

		return rows.Err()
	})

	return counts, err
}

// Cleanup removes at most batch terminal rows beyond the newest keep. Call
// it in a loop until it returns zero; never vacuum during an interaction.
func (s *Store) Cleanup(ctx context.Context, keep, batch int) (int, error) {
	if keep < 0 {
		keep = 0
	}

	if batch <= 0 {
		batch = 100
	}

	var removed int64

	err := s.do(ctx, func(ctx context.Context) error {
		res, err := s.db.ExecContext(ctx,
			`DELETE FROM jobs WHERE id IN (
				SELECT id FROM jobs
				WHERE state IN ('completed','failed','cancelled')
				ORDER BY updated_ms DESC, id DESC
				LIMIT ? OFFSET ?)`, batch, keep)
		if err != nil {
			return err
		}

		removed, err = res.RowsAffected()

		return err
	})

	return int(removed), err
}
