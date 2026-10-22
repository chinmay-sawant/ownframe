package storage

import (
	"context"
	"time"
)

// Aggregates returns downsampled buckets in bucket order.
func (s *Store) Aggregates(ctx context.Context, q Query) ([]Aggregate, error) {
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

	query := `SELECT metric, device, bucket_ns, min_value, max_value, avg_value, count
FROM metric_aggregate WHERE session_id = ?`
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
		query += " AND bucket_ns >= ?"
		args = append(args, q.Since.UnixNano())
	}
	if !q.Until.IsZero() {
		query += " AND bucket_ns <= ?"
		args = append(args, q.Until.UnixNano())
	}

	query += " ORDER BY bucket_ns LIMIT ?"
	args = append(args, limit)

	ctx, cancel := s.opCtx(ctx)
	defer cancel()

	rows, err := s.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer func() { _ = rows.Close() }()

	var out []Aggregate

	for rows.Next() {
		var (
			a  Aggregate
			ns int64
		)
		if err := rows.Scan(&a.Metric, &a.Device, &ns, &a.Min, &a.Max, &a.Avg, &a.Count); err != nil {
			return nil, err
		}
		a.Bucket = time.Unix(0, ns)
		out = append(out, a)
	}

	return out, rows.Err()
}
