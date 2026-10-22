package store

import (
	"context"
	"database/sql"

	"github.com/chinmay-sawant/ownframe/examples/tetris/game"
)

// TopScores lists live players' scores, highest first, ties broken by ID
// ascending. Dummy entries never appear here.
func (s *Store) TopScores(ctx context.Context, limit int) ([]game.Result, error) {
	return s.listScores(ctx, 0, limit)
}

// DummyScores lists the seeded demo entries with the same ordering.
func (s *Store) DummyScores(ctx context.Context, limit int) ([]game.Result, error) {
	return s.listScores(ctx, 1, limit)
}

// listScores reads one page of one ranking.
func (s *Store) listScores(ctx context.Context, dummy int, limit int) ([]game.Result, error) {
	limit = clampLimit(limit)

	var out []game.Result

	err := s.do(ctx, func(ctx context.Context, db *sql.DB) error {
		rows, err := db.QueryContext(ctx, `SELECT `+scoreColumns+
			` FROM scores WHERE dummy = ? ORDER BY score DESC, id ASC LIMIT ?`,
			dummy, limit)
		if err != nil {
			return err
		}

		defer func() { _ = rows.Close() }()
		for rows.Next() {
			r, err := scanScore(rows)
			if err != nil {
				return err
			}

			out = append(out, r)
		}

		return rows.Err()
	})

	return out, err
}
