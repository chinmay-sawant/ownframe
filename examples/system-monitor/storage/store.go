package storage

import (
	"database/sql"
	"sync/atomic"
	"time"
)

// Options configures OpenWithOptions. Zero fields take the defaults.
type Options struct {
	// Dir is the data directory. Empty uses DefaultDir.
	Dir string
	// BusyTimeout bounds a wait on a locked database.
	BusyTimeout time.Duration
	// OpTimeout bounds each statement when the caller's context has no
	// closer deadline.
	OpTimeout time.Duration
	// ForceRollback skips WAL and keeps the rollback journal, which is the
	// documented fallback when WAL is unavailable.
	ForceRollback bool
	// Temp opens an in-memory database for a clearly labeled temporary
	// mode. Nothing persists.
	Temp bool
}

const (
	defaultBusyTimeout = 5 * time.Second
	defaultOpTimeout   = 5 * time.Second
	pragmaTimeout      = 10 * time.Second
	migrateTimeout     = 30 * time.Second
	dbName             = "monitor.db"
	// storeDirPerm keeps data directories out of group and other access.
	storeDirPerm = 0o750
)

// Open and OpenWithOptions live in open.go.

// Store is one open database. Methods are synchronous and safe for
// concurrent use; the single connection serializes them, so a history read
// never sees a half-written batch.
type Store struct {
	db        *sql.DB
	path      string
	journal   string
	opTimeout time.Duration
	temp      bool
	closed    atomic.Bool
}

// Close, Path, Temp, and JournalMode live in store_access.go.
