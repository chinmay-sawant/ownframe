package store

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"io/fs"
	"strings"

	sqlite "modernc.org/sqlite"
)

func applyJournal(ctx context.Context, db *sql.DB) (string, error) {
	var mode string

	if err := db.QueryRowContext(ctx, "PRAGMA journal_mode=WAL").Scan(&mode); err != nil {
		if isReadOnly(err) {
			return "", wrapOpen(err)
		}

		return "", err
	}

	if strings.EqualFold(mode, "wal") {
		return "wal", nil
	}

	// A filesystem that cannot host WAL keeps a rollback journal instead.
	var fallback string

	if err := db.QueryRowContext(ctx, "PRAGMA journal_mode=DELETE").Scan(&fallback); err == nil {
		return strings.ToLower(fallback), nil
	}

	return strings.ToLower(mode), nil
}

func isReadOnly(err error) bool {
	msg := strings.ToLower(err.Error())
	if errors.Is(err, fs.ErrPermission) ||
		strings.Contains(msg, "readonly") ||
		strings.Contains(msg, "read-only") ||
		strings.Contains(msg, "permission denied") ||
		strings.Contains(msg, "access is denied") {
		return true
	}

	var se *sqlite.Error
	if errors.As(err, &se) {
		switch se.Code() {
		case sqlitePerm, sqliteReadonly, sqliteCantOpen: // SQLITE_PERM, SQLITE_READONLY, SQLITE_CANTOPEN
			return true
		}
	}

	return false
}

func wrapOpen(err error) error {
	if isReadOnly(err) {
		return ErrReadOnly
	}

	msg := strings.ToLower(err.Error())
	if strings.Contains(msg, "not a database") ||
		strings.Contains(msg, "malformed") ||
		strings.Contains(msg, "encrypted") {
		return fmt.Errorf("%w: %v", ErrCorrupt, err)
	}

	return err
}
