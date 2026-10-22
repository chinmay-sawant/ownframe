package storage

import (
	"context"
	"time"
)

// History returns raw rows in ID order, oldest first.
func (s *Store) History(ctx context.Context, q Query) ([]Point, error) {
	if err := s.ready(); err != nil {
		return nil, err
	}
	if q.SessionID == 0 {
		return nil, errNoSession
	}

	limit := q.Limit
	if limit <= 0 {
		limit = defaultHistoryLimit
	}

	query := `SELECT id, metric, device, ts_ns, mono_ns, value, valid
FROM metric_history WHERE session_id = ?`
	args := []any{q.SessionID}

	if q.Metric != "" {
		query += " AND metric = ?"
		args = append(args, string(q.Metric))
	}
	if q.Device != "" {
		query += " AND device = ?"
		args = append(args, q.Device)
	}
	if !q.Since.IsZero() {
		query += " AND ts_ns >= ?"
		args = append(args, q.Since.UnixNano())
	}
	if !q.Until.IsZero() {
		query += " AND ts_ns <= ?"
		args = append(args, q.Until.UnixNano())
	}
	if q.AfterID > 0 {
		query += " AND id > ?"
		args = append(args, q.AfterID)
	}

	query += " ORDER BY id LIMIT ?"
	args = append(args, limit)

	ctx, cancel := s.opCtx(ctx)
	defer cancel()

	rows, err := s.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer func() { _ = rows.Close() }()

	var out []Point

	for rows.Next() {
		var (
			p     Point
			ts    int64
			mono  int64
			valid int
		)
		if err := rows.Scan(&p.ID, &p.Metric, &p.Device, &ts, &mono, &p.Value, &valid); err != nil {
			return nil, err
		}
		p.At = time.Unix(0, ts)
		p.Mono = time.Duration(mono)
		p.Valid = valid != 0
		out = append(out, p)
	}

	return out, rows.Err()
}

// CountRows returns how many raw rows a session holds.
func (s *Store) CountRows(ctx context.Context, sessionID int64) (int64, error) {
	if err := s.ready(); err != nil {
		return 0, err
	}

	ctx, cancel := s.opCtx(ctx)
	defer cancel()

	var n int64
	if err := s.db.QueryRowContext(ctx,
		`SELECT COUNT(*) FROM metric_history WHERE session_id = ?`, sessionID).Scan(&n); err != nil {
		return 0, err
	}

	return n, nil
}
