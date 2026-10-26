package main

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"

	_ "modernc.org/sqlite"
)

// Store keeps the probe history in SQLite.
type Store struct {
	db *sql.DB
}

// OpenStore opens the database and applies the schema.
func OpenStore(path string) (*Store, error) {
	db, err := sql.Open("sqlite", path)
	if err != nil {
		return nil, fmt.Errorf("open sqlite: %w", err)
	}

	db.SetMaxOpenConns(1)
	db.SetConnMaxIdleTime(time.Minute)

	store := &Store{db: db}

	if err := store.migrate(context.Background()); err != nil {
		_ = db.Close()

		return nil, err
	}

	return store, nil
}

// Close closes the database.
func (s *Store) Close() error {
	return s.db.Close()
}

// migrate creates the one table and its indexes.
func (s *Store) migrate(ctx context.Context) error {
	const schema = `
CREATE TABLE IF NOT EXISTS probes (
	id       INTEGER PRIMARY KEY AUTOINCREMENT,
	service  TEXT NOT NULL,
	at       INTEGER NOT NULL,
	state    TEXT NOT NULL,
	status   INTEGER NOT NULL,
	total_us INTEGER NOT NULL,
	ttfb_us  INTEGER NOT NULL,
	err      TEXT NOT NULL DEFAULT '',
	attempts INTEGER NOT NULL DEFAULT 1
);
CREATE INDEX IF NOT EXISTS probes_service_at ON probes (service, at DESC);
CREATE INDEX IF NOT EXISTS probes_at ON probes (at DESC);
`

	if _, err := s.db.ExecContext(ctx, schema); err != nil {
		return fmt.Errorf("schema: %w", err)
	}

	return nil
}

// Record stores one probe result.
func (s *Store) Record(ctx context.Context, result Result) error {
	const insert = `
INSERT INTO probes (service, at, state, status, total_us, ttfb_us, err, attempts)
VALUES (?, ?, ?, ?, ?, ?, ?, ?)
`

	_, err := s.db.ExecContext(ctx, insert,
		result.Service,
		result.At.Unix(),
		result.State,
		result.Status,
		result.Total.Microseconds(),
		result.TTFB.Microseconds(),
		result.Err,
		result.Attempts,
	)
	if err != nil {
		return fmt.Errorf("insert probe: %w", err)
	}

	return nil
}

// Latest returns the newest row for each service.
func (s *Store) Latest(ctx context.Context) ([]Result, error) {
	const query = `
SELECT p.service, p.at, p.state, p.status, p.total_us, p.ttfb_us, p.err, p.attempts
FROM probes p
JOIN (
	SELECT service, MAX(at) AS at FROM probes GROUP BY service
) newest ON newest.service = p.service AND newest.at = p.at
ORDER BY p.service
`

	rows, err := s.db.QueryContext(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("latest: %w", err)
	}
	defer rows.Close()

	var out []Result

	for rows.Next() {
		var (
			result        Result
			at            int64
			total, ttfb   int64
		)

		if err := rows.Scan(&result.Service, &at, &result.State, &result.Status, &total, &ttfb, &result.Err, &result.Attempts); err != nil {
			return nil, fmt.Errorf("scan latest: %w", err)
		}

		result.At = time.Unix(at, 0).UTC()
		result.Total = time.Duration(total) * time.Microsecond
		result.TTFB = time.Duration(ttfb) * time.Microsecond
		out = append(out, result)
	}

	return out, rows.Err()
}

// History returns the newest checks for one service.
func (s *Store) History(ctx context.Context, service string, limit int) ([]Result, error) {
	const query = `
SELECT service, at, state, status, total_us, ttfb_us, err, attempts
FROM probes
WHERE service = ?
ORDER BY at DESC
LIMIT ?
`

	rows, err := s.db.QueryContext(ctx, query, service, limit)
	if err != nil {
		return nil, fmt.Errorf("history: %w", err)
	}
	defer rows.Close()

	var out []Result

	for rows.Next() {
		var (
			result      Result
			at          int64
			total, ttfb int64
		)

		if err := rows.Scan(&result.Service, &at, &result.State, &result.Status, &total, &ttfb, &result.Err, &result.Attempts); err != nil {
			return nil, fmt.Errorf("scan history: %w", err)
		}

		result.At = time.Unix(at, 0).UTC()
		result.Total = time.Duration(total) * time.Microsecond
		result.TTFB = time.Duration(ttfb) * time.Microsecond
		out = append(out, result)
	}

	return out, rows.Err()
}

// Prune drops rows older than the retention window.
func (s *Store) Prune(ctx context.Context, keep time.Duration) (int64, error) {
	cutoff := time.Now().Add(-keep).Unix()

	res, err := s.db.ExecContext(ctx, `DELETE FROM probes WHERE at < ?`, cutoff)
	if err != nil {
		return 0, fmt.Errorf("prune: %w", err)
	}

	return res.RowsAffected()
}

// ErrEmptyHistory reports a service that has never been checked.
var ErrEmptyHistory = errors.New("empty history")
