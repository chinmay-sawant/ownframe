package store

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"

	"github.com/chinmay-sawant/ownframe/examples/tetris/game"
)

// LoadSnapshot returns the stored snapshot, or ok=false when the slot is
// empty. A stored snapshot that no longer validates is rejected.
func (s *Store) LoadSnapshot(ctx context.Context) (game.Snapshot, bool, error) {
	var raw string

	err := s.do(ctx, func(ctx context.Context, db *sql.DB) error {
		return db.QueryRowContext(ctx,
			`SELECT payload FROM snapshot WHERE id = 1`).Scan(&raw)
	})
	if errors.Is(err, sql.ErrNoRows) {
		return game.Snapshot{}, false, nil
	}

	if err != nil {
		return game.Snapshot{}, false, err
	}

	var snap game.Snapshot
	if err := json.Unmarshal([]byte(raw), &snap); err != nil {
		return game.Snapshot{}, false, fmt.Errorf("%w: snapshot json: %w", ErrInvalid, err)
	}

	if err := snap.Validate(); err != nil {
		return game.Snapshot{}, false, fmt.Errorf("%w: %w", ErrInvalid, err)
	}

	return snap, true, nil
}

// ClearSnapshot empties the resume slot. The UI calls it on restart and
// on game over.
func (s *Store) ClearSnapshot(ctx context.Context) error {
	return s.doRetry(ctx, func(ctx context.Context, db *sql.DB) error {
		_, err := db.ExecContext(ctx, `DELETE FROM snapshot WHERE id = 1`)

		return err
	})
}
