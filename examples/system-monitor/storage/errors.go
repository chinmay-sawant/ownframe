package storage

import (
	"errors"
	"time"
)

// SchemaVersion is the schema this build writes. A database with a higher
// version belongs to a newer build and Open refuses it without modifying it.
const (
	SchemaVersion = 1
	// Retention bounds keep the default caps legible in one place; the Max
	// variables stay assignable so tests can lower them.
	defaultMaxRawRows       = 1_000_000
	defaultMaxAggregateRows = 500_000
)

// Retention and bounding policy for metric history. Raw rows live 24 hours,
// then Retain folds them into five-minute aggregates; aggregates live 30
// days. The row caps bound the tables even when the clock jumps or a recorder
// runs for weeks. The caps are variables so tests can lower them.
const (
	RawRetention       = 24 * time.Hour
	AggregateRetention = 30 * 24 * time.Hour
	AggregateBucket    = 5 * time.Minute
	BatchRows          = 5000
)

var (
	// MaxRawRows caps metric_history.
	MaxRawRows = int64(defaultMaxRawRows)
	// MaxAggregateRows caps metric_aggregate.
	MaxAggregateRows = int64(defaultMaxAggregateRows)
)

// ErrSchemaNewer reports a database written by a newer build. Open leaves the
// file untouched.
var ErrSchemaNewer = errors.New("system-monitor: database schema is newer than this build")

// ErrClosed reports work on a closed store.
var ErrClosed = errors.New("system-monitor: store is closed")
