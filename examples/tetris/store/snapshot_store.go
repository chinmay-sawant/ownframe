package store

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"time"

	"github.com/chinmay-sawant/ownframe/examples/tetris/game"
)

// SaveSnapshot atomically stores one validated resumable snapshot in the
// single slot. An invalid snapshot is rejected with ErrInvalid.
func (s *Store) SaveSnapshot(ctx context.Context, snap game.Snapshot) error {
	if err := snap.Validate(); err != nil {
		return fmt.Errorf("%w: %w", ErrInvalid, err)
	}

	data, err := json.Marshal(snap)
	if err != nil {
		return err
	}

	return s.doRetry(ctx, func(ctx context.Context, db *sql.DB) error {
		_, err := db.ExecContext(ctx, `
			INSERT INTO snapshot (id, game_id, payload, updated_at)
			VALUES (1, ?, ?, ?)
			ON CONFLICT (id) DO UPDATE SET
				game_id = excluded.game_id,
				payload = excluded.payload,
				updated_at = excluded.updated_at`,
			snap.ID, string(data), time.Now().UTC().Format(time.RFC3339Nano))

		return err
	})
}
