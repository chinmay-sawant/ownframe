package store

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"

	"github.com/chinmay-sawant/ownframe/examples/tetris/game"
)

// Replay returns the replay stored with a completed game's ID.
func (s *Store) Replay(ctx context.Context, id string) (game.Replay, error) {
	var (
		rep     game.Replay
		fixture string
		events  string
	)

	err := s.do(ctx, func(ctx context.Context, db *sql.DB) error {
		return db.QueryRowContext(ctx, `
			SELECT seed, ruleset, fixture_version, fixture, events
			FROM replays WHERE game_id = ?`, id).
			Scan(&rep.Seed, &rep.Ruleset, &rep.FixtureVersion, &fixture, &events)
	})
	if errors.Is(err, sql.ErrNoRows) {
		return game.Replay{}, fmt.Errorf("%w: replay %s", ErrNotFound, id)
	}

	if err != nil {
		return game.Replay{}, err
	}

	rep.Fixture = fixture
	if err := json.Unmarshal([]byte(events), &rep.Events); err != nil {
		return game.Replay{}, fmt.Errorf("%w: replay events: %w", ErrInvalid, err)
	}

	return rep, rep.Validate()
}

// Checkpoint runs a WAL checkpoint. On a rollback journal or an in-memory
// database it does nothing and returns nil. Run it from a background job,
// not from the UI loop.
func (s *Store) Checkpoint(ctx context.Context) error {
	if s.mode != "wal" {
		return nil
	}

	return s.do(ctx, func(ctx context.Context, db *sql.DB) error {
		_, err := db.ExecContext(ctx, "PRAGMA wal_checkpoint(TRUNCATE)")

		return err
	})
}

// Backup writes a consistent copy with VACUUM INTO, which is safe while
// the database is open. The destination must not exist.
func (s *Store) Backup(ctx context.Context, dest string) error {
	if dest == "" {
		return fmt.Errorf("%w: backup destination", ErrInvalid)
	}

	if err := os.MkdirAll(filepath.Dir(dest), dirPerm); err != nil {
		return err
	}

	return s.doRetry(ctx, func(ctx context.Context, db *sql.DB) error {
		_, err := db.ExecContext(ctx, `VACUUM INTO ?`, dest)

		return err
	})
}
