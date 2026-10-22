package store

import (
	"context"
	"database/sql"

	"github.com/chinmay-sawant/ownframe/examples/live-log-viewer/entry"
)

func pageTx(ctx context.Context, db *sql.DB, q Query, o PageOptions) (Page, error) {
	var p Page

	where, args := whereSQL(q)

	if q.MaxID == 0 {
		var hw int64

		if err := db.QueryRowContext(ctx,
			`SELECT COALESCE(MAX(id), 0) FROM entries`+where, args...).Scan(&hw); err != nil {
			return p, err
		}

		q.MaxID = entry.EntryID(hw)
	}

	p.HighWater = q.MaxID

	var oldest, newest int64

	if err := db.QueryRowContext(ctx,
		`SELECT COALESCE(MIN(id), 0), COALESCE(MAX(id), 0) FROM entries`+where,
		args...).Scan(&oldest, &newest); err != nil {
		return p, err
	}

	p.Oldest, p.Newest = entry.EntryID(oldest), entry.EntryID(newest)

	cond := where + " AND id <= ?"
	margs := append(append([]any(nil), args...), int64(q.MaxID))

	if q.Cursor > 0 {
		if o.Ascending {
			cond += " AND id > ?"
		} else {
			cond += " AND id < ?"
		}

		margs = append(margs, int64(q.Cursor))
	}

	order := " DESC"
	if o.Ascending {
		order = " ASC"
	}

	margs = append(margs, o.Limit+1)

	rows, err := db.QueryContext(ctx,
		entryColumns+cond+" ORDER BY id"+order+" LIMIT ?", margs...)
	if err != nil {
		return p, err
	}

	defer func() { _ = rows.Close() }()

	for rows.Next() {
		e, err := scanEntry(rows)
		if err != nil {
			return p, err
		}

		p.Entries = append(p.Entries, e)
	}

	if err := rows.Err(); err != nil {
		return p, err
	}

	if len(p.Entries) > o.Limit {
		p.Entries = p.Entries[:o.Limit]
		p.More = true
	}

	if len(p.Entries) == 0 && q.Cursor > 0 && oldest > 0 && q.Cursor <= entry.EntryID(oldest) {
		p.Expired = true
	}

	return p, nil
}
