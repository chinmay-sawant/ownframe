package store

import (
	"context"
	"database/sql"
	"errors"

	"github.com/chinmay-sawant/ownframe/examples/teams/activity"
)

// loadActivity reads the feed and the filter state.
func loadActivity(db *sql.DB) (activity.Data, error) {
	ctx := context.Background()
	rows, err := db.QueryContext(ctx, `SELECT id, actor, initials, color, kind, place, text,
		preview, time, unread, replies, replied FROM activity_items ORDER BY position`)
	if err != nil {
		return activity.Data{}, err
	}

	defer func() { _ = rows.Close() }()

	var items []activity.Item

	for rows.Next() {
		var it activity.Item

		if err := rows.Scan(&it.ID, &it.Actor, &it.Initials, &it.Color, &it.Kind,
			&it.Where, &it.Text, &it.Preview, &it.Time, &it.Unread, &it.Replies,
			&it.Replied); err != nil {
			return activity.Data{}, err
		}

		items = append(items, it)
	}

	if err := rows.Err(); err != nil {
		return activity.Data{}, err
	}

	var filter, active string

	err = db.QueryRowContext(ctx, `SELECT filter, active FROM activity_state WHERE id = 1`).
		Scan(&filter, &active)
	if err != nil && !errors.Is(err, sql.ErrNoRows) {
		return activity.Data{}, err
	}

	return activity.Data{Items: items, Filter: filter, Active: active}, nil
}
