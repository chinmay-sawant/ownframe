package store

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strconv"
)

// migrations holds one SQL script per version. Tests replace it to
// exercise a failed step.
var migrations = []string{schemaV1}

// migrate applies pending migrations, one transaction each. A newer
// schema is rejected before any write.
func migrate(ctx context.Context, db *sql.DB) error {
	v, err := storedVersion(ctx, db)
	if err != nil {
		return err
	}

	if v > SchemaVersion {
		return ErrNewerSchema
	}

	for next := v + 1; next <= len(migrations); next++ {
		if err := applyMigration(ctx, db, next, migrations[next-1]); err != nil {
			return fmt.Errorf("store: migration %d: %w", next, err)
		}
	}

	return nil
}

// storedVersion reads meta.schema_version, 0 if fresh.
func storedVersion(ctx context.Context, db *sql.DB) (int, error) {
	var raw string

	err := db.QueryRowContext(ctx,
		`SELECT value FROM meta WHERE key = 'schema_version'`).Scan(&raw)
	if errors.Is(err, sql.ErrNoRows) || isMissingTable(err) {
		return 0, nil
	}

	if err != nil {
		return 0, err
	}

	v, err := strconv.Atoi(raw)
	if err != nil {
		return 0, fmt.Errorf("store: schema version %q: %w", raw, err)
	}

	return v, nil
}

// applyMigration runs one script and records its version in one
// transaction, so an interrupted migration rolls back whole.
func applyMigration(ctx context.Context, db *sql.DB, version int, script string) error {
	tx, err := db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}

	defer func() { _ = tx.Rollback() }()
	if _, err := tx.ExecContext(ctx, script); err != nil {
		return err
	}

	_, err = tx.ExecContext(ctx, `
		INSERT INTO meta (key, value) VALUES ('schema_version', ?)
		ON CONFLICT (key) DO UPDATE SET value = excluded.value`,
		strconv.Itoa(version))
	if err != nil {
		return err
	}

	return tx.Commit()
}
