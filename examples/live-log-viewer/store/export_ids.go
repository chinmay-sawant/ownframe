package store

import (
	"context"
	"database/sql"
	"strings"

	"github.com/chinmay-sawant/ownframe/examples/live-log-viewer/entry"
)

func exportIDs(ctx context.Context, db *sql.DB, ids []entry.EntryID, add func(entry.Entry) bool) error {
	stop := false

	for start := 0; start < len(ids) && !stop; start += exportChunk {
		end := start + exportChunk
		if end > len(ids) {
			end = len(ids)
		}

		rows, err := queryIDs(ctx, db, ids[start:end])
		if err != nil {
			return err
		}

		for rows.Next() {
			e, err := scanEntry(rows)
			if err != nil {
				_ = rows.Close()

				return err
			}

			if !add(e) {
				stop = true

				break
			}
		}

		err = rows.Err()
		_ = rows.Close()

		if err != nil {
			return err
		}
	}

	return nil
}

func exportQuery(ctx context.Context, db *sql.DB, q Query, add func(entry.Entry) bool) error {
	for {
		p, err := pageTx(ctx, db, q, PageOptions{Limit: exportPageLimit, Ascending: true})
		if err != nil {
			return err
		}

		if q.MaxID == 0 {
			q.MaxID = p.HighWater
		}

		for _, e := range p.Entries {
			if !add(e) {
				return nil
			}
		}

		if len(p.Entries) == 0 || !p.More {
			return nil
		}

		q.Cursor = p.Entries[len(p.Entries)-1].ID
	}
}

func queryIDs(ctx context.Context, db *sql.DB, ids []entry.EntryID) (*sql.Rows, error) {
	args := make([]any, 0, len(ids))

	var b strings.Builder

	for i, id := range ids {
		if i > 0 {
			b.WriteString(", ")
		}

		b.WriteString("?")
		args = append(args, int64(id))
	}

	return db.QueryContext(ctx,
		entryColumns+` WHERE id IN (`+b.String()+`) ORDER BY id`, args...)
}
