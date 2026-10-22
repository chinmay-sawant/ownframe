package store

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/chinmay-sawant/ownframe/examples/live-log-viewer/entry"
)

func commitTx(ctx context.Context, db *sql.DB, src entry.SourceID, m CommitMeta, entries []entry.Entry) (int, error) {
	tx, err := db.BeginTx(ctx, nil)
	if err != nil {
		return 0, err
	}

	defer func() { _ = tx.Rollback() }()

	var maxSeq int64

	if err := tx.QueryRowContext(ctx,
		`SELECT COALESCE(MAX(source_seq), 0) FROM entries WHERE source_id = ?`,
		int64(src)).Scan(&maxSeq); err != nil {
		return 0, err
	}

	inserted, err := insertEntries(ctx, tx, entries, maxSeq)
	if err != nil {
		return 0, err
	}

	res, err := tx.ExecContext(ctx,
		`UPDATE sources SET
			identity   = COALESCE(NULLIF(?, ''), identity),
			head_hash  = COALESCE(NULLIF(?, 0), head_hash),
			head_len   = COALESCE(NULLIF(?, 0), head_len),
			generation = ?,
			position   = ?,
			size       = ?,
			state      = COALESCE(NULLIF(?, ''), state),
			lost       = lost + ?,
			updated_ns = ?
		 WHERE id = ?`,
		m.Identity, int64(m.HeadHash), m.HeadLen, maxGen(m), m.Position, m.Size, //nolint:gosec // reason: head hash round-trips through the signed column and back to uint64
		string(m.State), m.Lost, time.Now().UnixNano(), int64(src))
	if err != nil {
		return 0, err
	}

	if n, _ := res.RowsAffected(); n == 0 {
		return 0, fmt.Errorf("store: source %d is gone: %w", src, sql.ErrNoRows)
	}

	if err := tx.Commit(); err != nil {
		return 0, err
	}

	return inserted, nil
}

func maxGen(m CommitMeta) int64 {
	if m.Generation < 1 {
		return 1
	}

	return m.Generation
}
