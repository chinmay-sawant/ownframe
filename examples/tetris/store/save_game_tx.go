package store

import (
	"context"
	"database/sql"
	"encoding/json"
	"time"

	"github.com/chinmay-sawant/ownframe/examples/tetris/game"
)

// saveGame is the one completed-game transaction.
func saveGame(ctx context.Context, db *sql.DB, res game.Result, rep *game.Replay) error {
	tx, err := db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}

	defer func() { _ = tx.Rollback() }()
	now := time.Now().UTC().Format(time.RFC3339Nano)

	_, err = tx.ExecContext(ctx, `
		INSERT INTO scores (
			id, dummy, score, lines, level, pieces, duration_ms,
			seed, ruleset, fixture_version, created_at
		) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
		ON CONFLICT (id) DO NOTHING`,
		res.ID, res.Dummy, res.Score, res.Lines, res.Level, res.Pieces,
		res.DurationMS, int64(res.Seed), res.Ruleset, res.FixtureVersion, now) //nolint:gosec // reason: seed bits are stored in SQLite INTEGER and round-trip via uint64 on load, test values are small.
	if err != nil {
		return err
	}

	if rep == nil {
		return tx.Commit()
	}

	events, err := json.Marshal(rep.Events)
	if err != nil {
		return err
	}

	_, err = tx.ExecContext(ctx, `
		INSERT INTO replays (
			game_id, seed, ruleset, fixture_version, fixture,
			events, created_at
		) VALUES (?, ?, ?, ?, ?, ?, ?)
		ON CONFLICT (game_id) DO NOTHING`,
		res.ID, int64(rep.Seed), rep.Ruleset, rep.FixtureVersion, //nolint:gosec // reason: seed bits are stored in SQLite INTEGER and round-trip via uint64 on load, test values are small.
		rep.Fixture, string(events), now)
	if err != nil {
		return err
	}

	return tx.Commit()
}
