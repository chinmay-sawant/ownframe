package store

import (
	"context"
	"database/sql"
	"time"

	"github.com/chinmay-sawant/ownframe/examples/live-log-viewer/entry"
)

// Sessions lists every stored session in creation order.
func (s *Store) Sessions(ctx context.Context) ([]entry.Session, error) {
	return runJob(s, ctx, func(ctx context.Context, db *sql.DB) ([]entry.Session, error) {
		rows, err := db.QueryContext(ctx,
			`SELECT id, name, kind, created_ns FROM sessions ORDER BY id`)
		if err != nil {
			return nil, err
		}

		defer func() { _ = rows.Close() }()

		var out []entry.Session

		for rows.Next() {
			var (
				sess entry.Session
				ns   int64
			)

			if err := rows.Scan(&sess.ID, &sess.Name, &sess.Kind, &ns); err != nil {
				return nil, err
			}

			sess.Created = time.Unix(0, ns)
			out = append(out, sess)
		}

		return out, rows.Err()
	})
}

// EnsureSession returns the named session, creating it when missing.
func (s *Store) EnsureSession(ctx context.Context, name, kind string) (entry.Session, error) {
	return runJob(s, ctx, func(ctx context.Context, db *sql.DB) (entry.Session, error) {
		var zero entry.Session

		_, err := db.ExecContext(ctx,
			`INSERT INTO sessions(name, kind, created_ns) VALUES(?, ?, ?)
			 ON CONFLICT(name) DO NOTHING`,
			name, kind, time.Now().UnixNano())
		if err != nil {
			return zero, err
		}

		return sessionByName(ctx, db, name)
	})
}

// DeleteSession removes a session with its sources and entries.
func (s *Store) DeleteSession(ctx context.Context, id entry.SessionID) error {
	_, err := runJob(s, ctx, func(ctx context.Context, db *sql.DB) (int, error) {
		_, err := db.ExecContext(ctx, `DELETE FROM sessions WHERE id = ?`, int64(id))

		return 0, err
	})

	return err
}
