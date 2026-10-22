package storage

import (
	"context"
	"database/sql"
	"errors"
	"time"
)

// Setting is one saved preference, such as the theme or the selected mode.
type Setting struct {
	Key     string
	Value   string
	Updated time.Time
}

// Settings returns every setting, sorted by key.
func (s *Store) Settings(ctx context.Context) ([]Setting, error) {
	if err := s.ready(); err != nil {
		return nil, err
	}

	ctx, cancel := s.opCtx(ctx)
	defer cancel()

	rows, err := s.db.QueryContext(ctx, `SELECT key, value, updated_ns FROM settings ORDER BY key`)
	if err != nil {
		return nil, err
	}
	defer func() { _ = rows.Close() }()

	var out []Setting

	for rows.Next() {
		var (
			st Setting
			ns int64
		)
		if err := rows.Scan(&st.Key, &st.Value, &ns); err != nil {
			return nil, err
		}
		st.Updated = time.Unix(0, ns)
		out = append(out, st)
	}

	return out, rows.Err()
}

// Setting returns one value and whether it exists.
func (s *Store) Setting(ctx context.Context, key string) (string, bool, error) {
	if err := s.ready(); err != nil {
		return "", false, err
	}
	if key == "" {
		return "", false, errors.New("storage: empty setting key")
	}

	ctx, cancel := s.opCtx(ctx)
	defer cancel()

	var value string

	err := s.db.QueryRowContext(ctx, `SELECT value FROM settings WHERE key = ?`, key).Scan(&value)
	if errors.Is(err, sql.ErrNoRows) {
		return "", false, nil
	}
	if err != nil {
		return "", false, err
	}

	return value, true, nil
}
