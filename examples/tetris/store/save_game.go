package store

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/chinmay-sawant/ownframe/examples/tetris/game"
)

// SaveGame writes one completed game and its optional replay in one
// transaction. Re-inserting the same result ID changes nothing, so a
// retry after a failure cannot duplicate the score. Pass the same result
// again to retry.
func (s *Store) SaveGame(ctx context.Context, res game.Result, rep *game.Replay) error {
	if err := validateResult(res); err != nil {
		return err
	}

	if rep != nil {
		if err := rep.Validate(); err != nil {
			return fmt.Errorf("%w: %v", ErrInvalid, err)
		}
	}

	return s.doRetry(ctx, func(ctx context.Context, db *sql.DB) error {
		return saveGame(ctx, db, res, rep)
	})
}
