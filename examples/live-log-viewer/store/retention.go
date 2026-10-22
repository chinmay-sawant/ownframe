package store

import (
	"context"
	"database/sql"
	"time"

	"github.com/chinmay-sawant/ownframe/examples/live-log-viewer/entry"
)

// Retention bounds stored history. A zero field means unlimited.
type Retention struct {
	MaxRows  int64
	MaxBytes int64
	MaxAge   time.Duration
}

// DefaultRetention keeps a week of history, capped at 200,000 rows and
// 256 MiB of raw record bytes.
func DefaultRetention() Retention {
	return Retention{
		MaxRows:  retainRows,
		MaxBytes: retainBytes,
		MaxAge:   7 * 24 * time.Hour,
	}
}

// Pruned reports what one Prune call removed.
type Pruned struct {
	Entries int64
	Batches int
	Newest  entry.EntryID
}

// pruneBatches bounds one Prune call so maintenance never stalls an
// interaction. The next call continues where this one stopped.
const pruneBatches = 64

const pruneBatch = int64(1000)

// Prune deletes history in bounded batches, oldest first.
func (s *Store) Prune(ctx context.Context) (Pruned, error) {
	r := s.retention

	return runJobLong(s, ctx, func(ctx context.Context, db *sql.DB) (Pruned, error) {
		return pruneTx(ctx, db, r)
	})
}
