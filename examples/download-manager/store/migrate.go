package store

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
)

// migration is one versioned schema step.
type migration struct {
	version int
	script  string
}

// NewerSchemaError reports a newer database. Open does not touch it.
type NewerSchemaError struct {
	Found     int
	Supported int
}

func (e *NewerSchemaError) Error() string {
	return fmt.Sprintf("store: database schema v%d is newer than supported v%d",
		e.Found, e.Supported)
}

// ErrNewerSchema matches every NewerSchemaError.
var ErrNewerSchema = errors.New("store: newer schema")

// startup checks readability, rejects a newer schema, applies
// migrations, then picks WAL or the rollback journal.
func (s *Store) startup(ctx context.Context) error {
	var tables int

	if err := s.db.QueryRowContext(ctx,
		`SELECT count(*) FROM sqlite_master`).Scan(&tables); err != nil {
		return fmt.Errorf("store: unreadable database: %w", err)
	}

	version, hasMigrations, err := s.schemaVersion(ctx)
	if err != nil {
		return err
	}

	if version > len(migrations) {
		return fmt.Errorf("%w: %w", ErrNewerSchema, &NewerSchemaError{
			Found: version, Supported: len(migrations),
		})
	}

	from := version
	if !hasMigrations {
		from = 0
	}

	if err := s.apply(ctx, from); err != nil {
		return err
	}

	return s.enableJournal(ctx)
}

// schemaVersion reads the highest migration version. has is false when
// the table does not exist yet.
func (s *Store) schemaVersion(ctx context.Context) (version int, has bool, err error) {
	var name string

	err = s.db.QueryRowContext(ctx,
		`SELECT name FROM sqlite_master WHERE type = 'table' AND name = 'schema_migrations'`,
	).Scan(&name)

	if errors.Is(err, sql.ErrNoRows) {
		return 0, false, nil
	}

	if err != nil {
		return 0, false, err
	}

	var max sql.NullInt64

	err = s.db.QueryRowContext(ctx,
		`SELECT max(version) FROM schema_migrations`).Scan(&max)
	if err != nil {
		return 0, true, err
	}

	if max.Valid {
		version = int(max.Int64)
	}

	return version, true, nil
}
