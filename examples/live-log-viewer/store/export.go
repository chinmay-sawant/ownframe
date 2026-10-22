package store

import (
	"context"
	"database/sql"
	"strings"

	"github.com/chinmay-sawant/ownframe/examples/live-log-viewer/entry"
)

// ExportOptions bounds a selection export. With IDs set, exactly those rows
// are exported; otherwise Query selects them in ascending order.
type ExportOptions struct {
	Query      Query
	IDs        []entry.EntryID
	MaxEntries int
	MaxBytes   int
}

// Export is a bounded text export. Truncated reports that a limit stopped
// the selection early.
type Export struct {
	Data      []byte
	Entries   int
	Truncated bool
	Oldest    entry.EntryID
	Newest    entry.EntryID
}

const (
	exportMaxEntries = 20000
	exportMaxBytes   = 32 << 20
	exportChunk      = 400
)

// Export renders matching entries as plain text lines. The default bound is
// 5,000 entries or 8 MiB, whichever comes first.
func (s *Store) Export(ctx context.Context, o ExportOptions) (Export, error) {
	if o.MaxEntries <= 0 {
		o.MaxEntries = 5000
	}

	if o.MaxEntries > exportMaxEntries {
		o.MaxEntries = exportMaxEntries
	}

	if o.MaxBytes <= 0 {
		o.MaxBytes = exportDefaultCap
	}

	if o.MaxBytes > exportMaxBytes {
		o.MaxBytes = exportMaxBytes
	}

	return runJobLong(s, ctx, func(ctx context.Context, db *sql.DB) (Export, error) {
		return exportTx(ctx, db, o)
	})
}

func exportTx(ctx context.Context, db *sql.DB, o ExportOptions) (Export, error) {
	var (
		out Export
		b   strings.Builder
	)

	add := func(e entry.Entry) bool {
		line := exportLine(e)
		if out.Entries >= o.MaxEntries || b.Len()+len(line) > o.MaxBytes {
			out.Truncated = true

			return false
		}

		if out.Entries == 0 {
			out.Oldest = e.ID
		}

		out.Newest = e.ID
		out.Entries++
		b.WriteString(line)

		return true
	}

	if len(o.IDs) > 0 {
		if err := exportIDs(ctx, db, o.IDs, add); err != nil {
			return out, err
		}
	} else if err := exportQuery(ctx, db, o.Query, add); err != nil {
		return out, err
	}

	out.Data = []byte(b.String())

	return out, nil
}
