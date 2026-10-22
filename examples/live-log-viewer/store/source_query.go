package store

import (
	"context"
	"database/sql"
	"time"

	"github.com/chinmay-sawant/ownframe/examples/live-log-viewer/entry"
)

// Sources lists the sources of one session.
func (s *Store) Sources(ctx context.Context, session entry.SessionID) ([]entry.Source, error) {
	return runJob(s, ctx, func(ctx context.Context, db *sql.DB) ([]entry.Source, error) {
		rows, err := db.QueryContext(ctx, sourceColumns+
			` WHERE session_id = ? ORDER BY id`, int64(session))
		if err != nil {
			return nil, err
		}

		defer func() { _ = rows.Close() }()

		var out []entry.Source

		for rows.Next() {
			src, err := scanSource(rows)
			if err != nil {
				return nil, err
			}

			out = append(out, src)
		}

		return out, rows.Err()
	})
}

// Source loads one source by id.
func (s *Store) Source(ctx context.Context, id entry.SourceID) (entry.Source, error) {
	return runJob(s, ctx, func(ctx context.Context, db *sql.DB) (entry.Source, error) {
		row := db.QueryRowContext(ctx, sourceColumns+` WHERE id = ?`, int64(id))

		return scanSource(row)
	})
}

// DeleteSource removes a source with its entries.
func (s *Store) DeleteSource(ctx context.Context, id entry.SourceID) error {
	_, err := runJob(s, ctx, func(ctx context.Context, db *sql.DB) (int, error) {
		_, err := db.ExecContext(ctx, `DELETE FROM sources WHERE id = ?`, int64(id))

		return 0, err
	})

	return err
}

// SetSourceState records what a source is doing without touching its
// checkpoint.
func (s *Store) SetSourceState(ctx context.Context, id entry.SourceID, state entry.State) error {
	_, err := runJob(s, ctx, func(ctx context.Context, db *sql.DB) (int, error) {
		_, err := db.ExecContext(ctx,
			`UPDATE sources SET state = ?, updated_ns = ? WHERE id = ?`,
			string(state), time.Now().UnixNano(), int64(id))

		return 0, err
	})

	return err
}
