package store

import (
	"context"
	"database/sql"
	"time"

	"github.com/chinmay-sawant/ownframe/examples/live-log-viewer/entry"
)

// Query selects entries. Nil fields match everything. Cursor is a keyset
// bound: with PageOptions.Ascending false it returns entries with a smaller
// ID, with Ascending true a larger one. MaxID freezes the view at a
// high-water mark; zero means live and resolves to the newest match.
type Query struct {
	Session     *entry.SessionID
	Source      *entry.SourceID
	MinSeverity *entry.Severity
	Text        string
	Since       *time.Time
	Until       *time.Time
	Cursor      entry.EntryID
	MaxID       entry.EntryID
}

// PageOptions controls one page.
type PageOptions struct {
	Limit     int
	Ascending bool
}

// Page is one bounded result window. HighWater is the resolved upper ID;
// pass it back in Query.MaxID to keep browsing the same frozen view.
// Expired reports that the cursor fell below retained history.
type Page struct {
	Entries   []entry.Entry
	HighWater entry.EntryID
	Oldest    entry.EntryID
	Newest    entry.EntryID
	More      bool
	Expired   bool
}

// Page returns at most Limit entries ordered by ID. The default limit is
// 200 and the cap is 500.
func (s *Store) Page(ctx context.Context, q Query, o PageOptions) (Page, error) {
	if o.Limit < 1 {
		o.Limit = entry.DefaultPolicy().PageLimit
	}

	if o.Limit > pageCap {
		o.Limit = pageCap
	}

	return runJob(s, ctx, func(ctx context.Context, db *sql.DB) (Page, error) {
		return pageTx(ctx, db, q, o)
	})
}

// Count returns the number of matching entries. With Cursor > 0 it counts
// only entries above the cursor, which is what an unread badge needs.
func (s *Store) Count(ctx context.Context, q Query) (int64, error) {
	return runJob(s, ctx, func(ctx context.Context, db *sql.DB) (int64, error) {
		return countTx(ctx, db, q)
	})
}
