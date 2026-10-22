package store

import (
	"database/sql"

	"github.com/chinmay-sawant/ownframe/examples/teams/activity"
)

// saveActivity replaces the feed with d.
func saveActivity(tx *sql.Tx, d activity.Data) error {
	if _, err := tx.Exec(`DELETE FROM activity_items`); err != nil {
		return err
	}

	for i, it := range d.Items {
		_, err := tx.Exec(`INSERT INTO activity_items (id, position, actor, initials,
			color, kind, place, text, preview, time, unread, replies, replied)
			VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`,
			it.ID, i, it.Actor, it.Initials, it.Color, it.Kind, it.Where, it.Text,
			it.Preview, it.Time, it.Unread, it.Replies, it.Replied)
		if err != nil {
			return err
		}
	}

	_, err := tx.Exec(`INSERT OR REPLACE INTO activity_state (id, filter, active)
		VALUES (1, ?, ?)`, d.Filter, d.Active)

	return err
}

// loadActivity reads the feed and the filter state.
func loadActivity(db *sql.DB) (activity.Data, error) {
	rows, err := db.Query(`SELECT id, actor, initials, color, kind, place, text,
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

	err = db.QueryRow(`SELECT filter, active FROM activity_state WHERE id = 1`).
		Scan(&filter, &active)
	if err != nil && err != sql.ErrNoRows {
		return activity.Data{}, err
	}

	return activity.Data{Items: items, Filter: filter, Active: active}, nil
}
