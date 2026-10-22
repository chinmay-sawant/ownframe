package store

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"time"
)

// maxSettingsJSON bounds a stored settings document.
const maxSettingsJSON = 4096

// SaveSettings stores the control configuration.
func (s *Store) SaveSettings(ctx context.Context, set Settings) error {
	data, err := json.Marshal(set)
	if err != nil {
		return fmt.Errorf("%w: settings: %w", ErrInvalid, err)
	}

	if len(data) > maxSettingsJSON {
		return fmt.Errorf("%w: settings too large", ErrInvalid)
	}

	return s.doRetry(ctx, func(ctx context.Context, db *sql.DB) error {
		_, err := db.ExecContext(ctx, `
			INSERT INTO settings (id, json, updated_at) VALUES (1, ?, ?)
			ON CONFLICT (id) DO UPDATE SET
				json = excluded.json,
				updated_at = excluded.updated_at`,
			string(data), time.Now().UTC().Format(time.RFC3339Nano))

		return err
	})
}

// Settings loads the stored configuration, or defaults when none exists.
func (s *Store) Settings(ctx context.Context) (Settings, error) {
	var raw string

	err := s.do(ctx, func(ctx context.Context, db *sql.DB) error {
		return db.QueryRowContext(ctx,
			`SELECT json FROM settings WHERE id = 1`).Scan(&raw)
	})
	if errors.Is(err, sql.ErrNoRows) {
		return DefaultSettings(), nil
	}

	if err != nil {
		return Settings{}, err
	}

	var set Settings
	if err := json.Unmarshal([]byte(raw), &set); err != nil {
		return Settings{}, fmt.Errorf("%w: settings json: %w", ErrInvalid, err)
	}

	return set, nil
}
