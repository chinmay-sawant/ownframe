package storage

import (
	"context"
	"database/sql"
	"errors"
)

// sessionsQuery is the shared read for Sessions and Session.
const sessionsQuery = `
SELECT s.id, s.name, s.source, s.mode, s.state, s.note,
       s.started_ns, s.ended_ns, s.sampling_ms,
       (SELECT COUNT(*) FROM metric_history h WHERE h.session_id = s.id)
FROM sessions s`

// scanSession is in sessions_scan.go.

// Sessions returns the most recent sessions first. limit <= 0 returns all.
func (s *Store) Sessions(ctx context.Context, limit int) ([]Session, error) {
	if err := s.ready(); err != nil {
		return nil, err
	}

	ctx, cancel := s.opCtx(ctx)
	defer cancel()

	query := sessionsQuery + ` ORDER BY s.id DESC`
	args := []any{}

	if limit > 0 {
		query += ` LIMIT ?`
		args = append(args, limit)
	}

	rows, err := s.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer func() { _ = rows.Close() }()

	var out []Session

	for rows.Next() {
		sess, err := scanSession(rows)
		if err != nil {
			return nil, err
		}
		out = append(out, sess)
	}

	return out, rows.Err()
}

// Session returns one session by ID.
func (s *Store) Session(ctx context.Context, id int64) (Session, bool, error) {
	if err := s.ready(); err != nil {
		return Session{}, false, err
	}

	ctx, cancel := s.opCtx(ctx)
	defer cancel()

	sess, err := scanSession(s.db.QueryRowContext(ctx, sessionsQuery+` WHERE s.id = ?`, id))
	if errors.Is(err, sql.ErrNoRows) {
		return Session{}, false, nil
	}
	if err != nil {
		return Session{}, false, err
	}

	return sess, true, nil
}
