package storage

import (
	"context"
	"database/sql"
)

// Append writes metric rows for a session. A duplicate row (same session,
// metric, device, and timestamp) is ignored rather than overwriting, so a
// retried batch cannot corrupt history. The second return counts rows the
// store kept.
func (s *Store) Append(ctx context.Context, sessionID int64, rows []Row) (int, error) {
	if err := s.ready(); err != nil {
		return 0, err
	}
	if len(rows) == 0 {
		return 0, nil
	}

	ctx, cancel := s.opCtx(ctx)
	defer cancel()

	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return 0, err
	}
	defer func() { _ = tx.Rollback() }()

	kept, err := appendTx(ctx, tx, sessionID, rows)
	if err != nil {
		return 0, err
	}

	return kept, tx.Commit()
}

// appendTx writes rows inside an existing transaction. Rows are bounded by
// the caller: the recorder sends one sample at a time and the seeder sends one
// fixture at a time.
func appendTx(ctx context.Context, tx *sql.Tx, sessionID int64, rows []Row) (int, error) {
	stmt, err := tx.PrepareContext(ctx, `
INSERT OR IGNORE INTO metric_history(session_id, metric, device, ts_ns, mono_ns, value, valid)
VALUES(?, ?, ?, ?, ?, ?, ?)`)
	if err != nil {
		return 0, err
	}
	defer func() { _ = stmt.Close() }()

	kept := 0

	for _, r := range rows {
		if r.Metric == "" {
			continue
		}

		res, err := stmt.ExecContext(ctx, sessionID, string(r.Metric), r.Device,
			r.At.UnixNano(), int64(r.Mono), r.Value, boolInt(r.Valid))
		if err != nil {
			return kept, err
		}

		if n, _ := res.RowsAffected(); n > 0 {
			kept++
		}
	}

	return kept, nil
}
