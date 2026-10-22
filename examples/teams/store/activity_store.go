package store

import (
	"context"
	"database/sql"

	"github.com/chinmay-sawant/ownframe/examples/teams/activity"
)

// saveActivity replaces the feed with d.
func saveActivity(tx *sql.Tx, d activity.Data) error {
	ctx := context.Background()
	if _, err := tx.ExecContext(ctx, `DELETE FROM activity_items`); err != nil {
		return err
	}

	for i, it := range d.Items {
		_, err := tx.ExecContext(ctx, `INSERT INTO activity_items (id, position, actor, initials,
			color, kind, place, text, preview, time, unread, replies, replied)
			VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`,
			it.ID, i, it.Actor, it.Initials, it.Color, it.Kind, it.Where, it.Text,
			it.Preview, it.Time, it.Unread, it.Replies, it.Replied)
		if err != nil {
			return err
		}
	}

	_, err := tx.ExecContext(ctx, `INSERT OR REPLACE INTO activity_state (id, filter, active)
		VALUES (1, ?, ?)`, d.Filter, d.Active)

	return err
}
