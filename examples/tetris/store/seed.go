package store

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strconv"
	"time"
)

// seedIfNeeded inserts the demo scores once, guarded by a seed version
// marker. Reopening leaves an existing marker and any user edits alone.
func seedIfNeeded(ctx context.Context, db *sql.DB) error {
	var raw string

	err := db.QueryRowContext(ctx,
		`SELECT value FROM meta WHERE key = 'seed_version'`).Scan(&raw)
	if err == nil {
		if v, convErr := strconv.Atoi(raw); convErr == nil && v >= SeedVersion {
			return nil
		}
	} else if !errors.Is(err, sql.ErrNoRows) && !isMissingTable(err) {
		return err
	}

	return applySeed(ctx, db)
}

// applySeed writes the dummy scores and the marker in one transaction.
func applySeed(ctx context.Context, db *sql.DB) error {
	tx, err := db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}

	defer func() { _ = tx.Rollback() }()
	for i, d := range dummyScores {
		id := fmt.Sprintf("dummy-%04d", i+1)
		day := time.Date(2026, time.October, d.day, 12, 0, 0, 0, time.UTC)

		_, err := tx.ExecContext(ctx, `
			INSERT INTO scores (
				id, dummy, score, lines, level, pieces, duration_ms,
				seed, ruleset, fixture_version, created_at
			) VALUES (?, 1, ?, ?, ?, ?, ?, ?, ?, ?, ?)
			ON CONFLICT (id) DO NOTHING`,
			id, d.score, d.lines, d.level, d.pieces,
			scorePerLine*d.lines, int64(d.seed), rulesetName, fixtureVersion, //nolint:gosec // reason: seed bits are stored in SQLite INTEGER and round-trip via uint64 on load, fixture values are small.
			day.Format(time.RFC3339))
		if err != nil {
			return err
		}
	}

	_, err = tx.ExecContext(ctx, `
		INSERT INTO meta (key, value) VALUES ('seed_version', ?)
		ON CONFLICT (key) DO UPDATE SET value = excluded.value`,
		strconv.Itoa(SeedVersion))
	if err != nil {
		return err
	}

	return tx.Commit()
}
